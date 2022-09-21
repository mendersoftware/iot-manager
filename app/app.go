// Copyright 2022 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package app

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/mendersoftware/go-lib-micro/identity"
	"github.com/mendersoftware/go-lib-micro/log"

	"github.com/mendersoftware/iot-manager/client"
	"github.com/mendersoftware/iot-manager/client/devauth"
	"github.com/mendersoftware/iot-manager/client/iotcore"
	"github.com/mendersoftware/iot-manager/client/iothub"
	"github.com/mendersoftware/iot-manager/client/workflows"
	"github.com/mendersoftware/iot-manager/model"
	"github.com/mendersoftware/iot-manager/store"
)

var (
	ErrIntegrationNotFound = errors.New("integration not found")
	ErrIntegrationExists   = errors.New("integration already exists")
	ErrUnknownIntegration  = errors.New("unknown integration provider")
	ErrNoCredentials       = errors.New("no connection string or credentials " +
		"configured for the tenant")
	ErrNoDeviceConnectionString = errors.New("device has no connection string")

	ErrDeviceAlreadyExists     = errors.New("device already exists")
	ErrDeviceNotFound          = errors.New("device not found")
	ErrDeviceStateConflict     = errors.New("conflict when updating the device state")
	ErrCannotRemoveIntegration = errors.New("cannot remove integration in use by devices")
)

const (
	confKeyPrimaryKey     = "azureConnectionString"
	confKeyAWSCertificate = "awsCertificate"
	confKeyAWSPrivateKey  = "awsPrivateKey"
	confKeyAWSEndpoint    = "awsEndpoint"
)

// App interface describes app objects
//
//nolint:lll
//go:generate ../utils/mockgen.sh
type App interface {
	WithIoTCore(client iotcore.Client) App
	WithIoTHub(client iothub.Client) App
	HealthCheck(context.Context) error
	GetDeviceIntegrations(context.Context, string) ([]model.Integration, error)
	GetIntegrations(context.Context) ([]model.Integration, error)
	GetIntegrationById(context.Context, uuid.UUID) (*model.Integration, error)
	CreateIntegration(context.Context, model.Integration) (*model.Integration, error)
	SetDeviceStatus(context.Context, string, model.Status) error
	SetIntegrationCredentials(context.Context, uuid.UUID, model.Credentials) error
	RemoveIntegration(context.Context, uuid.UUID) error
	GetDevice(context.Context, string) (*model.Device, error)
	GetDeviceStateIntegration(context.Context, string, uuid.UUID) (*model.DeviceState, error)
	SetDeviceStateIntegration(context.Context, string, uuid.UUID, *model.DeviceState) (*model.DeviceState, error)
	GetDeviceStateIoTHub(context.Context, string, *model.Integration) (*model.DeviceState, error)
	SetDeviceStateIoTHub(context.Context, string, *model.Integration, *model.DeviceState) (*model.DeviceState, error)
	GetDeviceStateIoTCore(context.Context, string, *model.Integration) (*model.DeviceState, error)
	SetDeviceStateIoTCore(context.Context, string, *model.Integration, *model.DeviceState) (*model.DeviceState, error)
	ProvisionDevice(context.Context, model.DeviceEvent) error
	DecommissionDevice(context.Context, string) error

	SyncDevices(context.Context, int, bool) error

	GetEvents(ctx context.Context, filter model.EventsFilter) ([]model.Event, error)
}

// app is an app object
type app struct {
	store         store.DataStore
	iothubClient  iothub.Client
	iotcoreClient iotcore.Client
	wf            workflows.Client
	devauth       devauth.Client
	httpClient    *http.Client
}

// NewApp initialize a new iot-manager App
func New(ds store.DataStore, wf workflows.Client, da devauth.Client) App {
	c := client.New()
	hubClient := iothub.NewClient(
		iothub.NewOptions().SetClient(c),
	)
	return &app{
		store:        ds,
		wf:           wf,
		devauth:      da,
		iothubClient: hubClient,
		httpClient:   c,
	}
}

// WithIoTHub sets the IoT Hub client
func (a *app) WithIoTHub(client iothub.Client) App {
	a.iothubClient = client
	return a
}

// WithIoTCore sets the IoT Core client
func (a *app) WithIoTCore(client iotcore.Client) App {
	a.iotcoreClient = client
	return a
}

// HealthCheck performs a health check and returns an error if it fails
func (a *app) HealthCheck(ctx context.Context) error {
	return a.store.Ping(ctx)
}

func (a *app) GetIntegrations(ctx context.Context) ([]model.Integration, error) {
	return a.store.GetIntegrations(ctx, model.IntegrationFilter{})
}

func (a *app) GetIntegrationById(ctx context.Context, id uuid.UUID) (*model.Integration, error) {
	integration, err := a.store.GetIntegrationById(ctx, id)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case store.ErrObjectNotFound:
			return nil, ErrIntegrationNotFound
		default:
			return nil, err
		}
	}
	return integration, err
}

func (a *app) CreateIntegration(
	ctx context.Context,
	integration model.Integration,
) (*model.Integration, error) {
	result, err := a.store.CreateIntegration(ctx, integration)
	if err == store.ErrObjectExists {
		return nil, ErrIntegrationExists
	}
	return result, err
}

func (a *app) SetIntegrationCredentials(
	ctx context.Context,
	integrationID uuid.UUID,
	credentials model.Credentials,
) error {
	err := a.store.SetIntegrationCredentials(ctx, integrationID, credentials)
	if err != nil {
		switch cause := errors.Cause(err); cause {
		case store.ErrObjectNotFound:
			return ErrIntegrationNotFound
		default:
			return err
		}
	}
	return err
}

func (a *app) RemoveIntegration(
	ctx context.Context,
	integrationID uuid.UUID,
) error {
	itg, err := a.store.GetIntegrationById(ctx, integrationID)
	if err == nil {
		if itg.Provider != model.ProviderWebhook {
			// check if there are any devices with given integration enabled
			devicesExist, err := a.store.
				DoDevicesExistByIntegrationID(ctx, integrationID)
			if err != nil {
				return err
			}
			if devicesExist {
				return ErrCannotRemoveIntegration
			}
		}
		err = a.store.RemoveIntegration(ctx, integrationID)
	}
	if errors.Is(err, store.ErrObjectNotFound) {
		return ErrIntegrationNotFound
	}
	return err
}

func (a *app) GetDeviceIntegrations(
	ctx context.Context,
	deviceID string,
) ([]model.Integration, error) {
	device, err := a.store.GetDevice(ctx, deviceID)
	if err != nil {
		if err == store.ErrObjectNotFound {
			return nil, ErrDeviceNotFound
		}
		return nil, errors.Wrap(err, "app: failed to get device integrations")
	}
	if len(device.IntegrationIDs) > 0 {
		integrations, err := a.store.GetIntegrations(ctx,
			model.IntegrationFilter{IDs: device.IntegrationIDs},
		)
		return integrations, errors.Wrap(err, "app: failed to get device integrations")
	}
	return []model.Integration{}, nil
}

func (a *app) SetDeviceStatus(ctx context.Context, deviceID string, status model.Status) error {
	integrations, err := a.store.GetIntegrations(ctx, model.IntegrationFilter{})
	if err != nil {
		return errors.Wrap(err, "failed to retrieve integrations")
	}
	event := model.Event{
		WebhookEvent: model.WebhookEvent{
			ID:   uuid.New(),
			Type: model.EventTypeDeviceStatusChanged,
			Data: model.DeviceEvent{
				ID:     deviceID,
				Status: status,
			},
			EventTS: time.Now(),
		},
		DeliveryStatus: make([]model.DeliveryStatus, 0, len(integrations)),
	}

	var (
		ok       bool
		errStack model.ErrorStack
		device   = newDevice(deviceID, a.store)
	)
	for _, integration := range integrations {
		deliver := model.DeliveryStatus{
			IntegrationID: integration.ID,
			Success:       true,
		}
		switch integration.Provider {
		case model.ProviderIoTHub:
			ok, err = device.HasIntegration(ctx, integration.ID)
			if err != nil {
				break // switch
			} else if !ok {
				continue // loop
			}
			err = a.setDeviceStatusIoTHub(ctx, deviceID, status, integration)

		case model.ProviderIoTCore:
			ok, err = device.HasIntegration(ctx, integration.ID)
			if err != nil {
				break // switch
			} else if !ok {
				continue // loop
			}
			err = a.setDeviceStatusIoTCore(ctx, deviceID, status, integration)

		case model.ProviderWebhook:
			var (
				req *http.Request
				rsp *http.Response
			)
			req, err = client.NewWebhookRequest(ctx,
				&integration.Credentials,
				event.WebhookEvent)
			if err != nil {
				break // switch
			}
			rsp, err = a.httpClient.Do(req)
			if err != nil {
				break // switch
			}
			deliver.StatusCode = &rsp.StatusCode
			if rsp.StatusCode >= 300 {
				err = client.NewHTTPError(rsp.StatusCode)
			}
			_ = rsp.Body.Close()

		default:
			continue
		}
		if err != nil {
			var httpError client.HTTPError
			if errors.As(err, &httpError) {
				errCode := httpError.Code()
				deliver.StatusCode = &errCode
			}
			deliver.Success = false
			deliver.Error = err.Error()
			_ = errStack.Push(err)
		}
		event.DeliveryStatus = append(event.DeliveryStatus, deliver)
	}
	err = a.store.SaveEvent(ctx, event)
	if errStack != nil {
		if err != nil {
			err = errors.WithMessage(err, errStack.Error())
		} else {
			err = errStack
		}
	}
	return err
}

func (a *app) ProvisionDevice(
	ctx context.Context,
	device model.DeviceEvent,
) error {
	integrations, err := a.GetIntegrations(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve integrations")
	}
	var errStack model.ErrorStack
	event := model.Event{
		WebhookEvent: model.WebhookEvent{
			ID:      uuid.New(),
			Type:    model.EventTypeDeviceProvisioned,
			Data:    device,
			EventTS: time.Now(),
		},
		DeliveryStatus: make([]model.DeliveryStatus, 0, len(integrations)),
	}
	integrationIDs := make([]uuid.UUID, 0, len(integrations))
	for _, integration := range integrations {
		deliver := model.DeliveryStatus{
			IntegrationID: integration.ID,
			Success:       true,
		}
		switch integration.Provider {
		case model.ProviderIoTHub:
			err = a.provisionIoTHubDevice(ctx, device.ID, integration)
			integrationIDs = append(integrationIDs, integration.ID)
		case model.ProviderIoTCore:
			err = a.provisionIoTCoreDevice(ctx, device.ID, integration, &iotcore.Device{
				Status: iotcore.StatusEnabled,
			})
			integrationIDs = append(integrationIDs, integration.ID)
		case model.ProviderWebhook:
			var (
				req *http.Request
				rsp *http.Response
			)
			req, err = client.NewWebhookRequest(ctx,
				&integration.Credentials,
				event.WebhookEvent)
			if err != nil {
				break // switch
			}
			rsp, err = a.httpClient.Do(req)
			if err != nil {
				break // switch
			}
			deliver.StatusCode = &rsp.StatusCode
			if rsp.StatusCode >= 300 {
				err = client.NewHTTPError(rsp.StatusCode)
			}
			_ = rsp.Body.Close()

		default:
			continue
		}
		if err != nil {
			var httpError client.HTTPError
			if errors.As(err, &httpError) {
				errCode := httpError.Code()
				deliver.StatusCode = &errCode
			}
			deliver.Success = false
			deliver.Error = err.Error()
			_ = errStack.Push(err)
		}
		event.DeliveryStatus = append(event.DeliveryStatus, deliver)
	}
	_, err = a.store.UpsertDeviceIntegrations(ctx, device.ID, integrationIDs)
	errSave := a.store.SaveEvent(ctx, event)
	if errSave != nil {
		_ = errStack.Push(errSave)
	}
	if errStack != nil {
		if err != nil {
			err = errors.Wrap(err, errStack.Error())
		} else {
			err = errStack
		}
	}
	return err
}

func (a *app) syncBatch(
	ctx context.Context,
	devices []model.Device,
	integCache map[uuid.UUID]*model.Integration,
	failEarly bool,
) error {
	var err error
	l := log.FromContext(ctx)

	deviceMap := make(map[uuid.UUID][]string, len(integCache))
	for _, dev := range devices {
		for _, id := range dev.IntegrationIDs {
			deviceMap[id] = append(deviceMap[id], dev.ID)
		}
	}

	for integID, deviceIDs := range deviceMap {
		integration, ok := integCache[integID]
		if !ok {
			// (Data race) Try again to fetch the integration
			integration, err = a.store.GetIntegrationById(ctx, integID)
			if err != nil {
				if err == store.ErrObjectNotFound {
					integCache[integID] = nil
					continue
				}
				err = errors.Wrap(err, "failed to retrieve device integration")
				if failEarly {
					return err
				}
				l.Errorf("failed to get device integration: %s", err)
				continue
			} else {
				integCache[integID] = integration
			}
		}
		if integration == nil {
			// Should not occur, but is not impossible since mongo client
			// caches batches of results.
			_, err := a.store.RemoveDevicesFromIntegration(ctx, integID)
			if err != nil {
				err = errors.Wrap(err, "failed to remove integration from devices")
				if failEarly {
					return err
				}
				l.Error(err)
			}
			continue
		}

		switch integration.Provider {
		case model.ProviderIoTHub:
			err := a.syncIoTHubDevices(ctx, deviceIDs, *integration, failEarly)
			if err != nil {
				if failEarly {
					return err
				}
				l.Error(err)
			}
		case model.ProviderIoTCore:
			err := a.syncIoTCoreDevices(ctx, deviceIDs, *integration, failEarly)
			if err != nil {
				if failEarly {
					return err
				}
				l.Error(err)
			}
		default:
		}
	}

	return nil
}

func (a app) syncCacheIntegrations(ctx context.Context) (map[uuid.UUID]*model.Integration, error) {
	// NOTE At the time of writing this, we don't allow more than one
	//      integration per tenant so this const doesn't matter.
	// TODO Will we need a more sophisticated cache data structure?
	const MaxIntegrationsToCache = 20
	// Cache integrations for the given tenant
	integrations, err := a.store.GetIntegrations(
		ctx, model.IntegrationFilter{Limit: MaxIntegrationsToCache},
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get integrations for tenant")
	}
	integCache := make(map[uuid.UUID]*model.Integration, len(integrations))
	for i := range integrations {
		integCache[integrations[i].ID] = &integrations[i]
	}
	return integCache, nil
}

func (a *app) SyncDevices(
	ctx context.Context,
	batchSize int,
	failEarly bool,
) error {
	type DeviceWithTenantID struct {
		model.Device `bson:",inline"`
		TenantID     string `bson:"tenant_id"`
	}
	iter, err := a.store.GetAllDevices(ctx)
	if err != nil {
		return err
	}
	defer iter.Close(ctx)

	var (
		deviceBatch        = make([]model.Device, 0, batchSize)
		tenantID    string = ""
		integCache  map[uuid.UUID]*model.Integration
	)
	tCtx := identity.WithContext(ctx, &identity.Identity{
		Tenant: tenantID,
	})
	integCache, err = a.syncCacheIntegrations(tCtx)
	if err != nil {
		return err
	}
	for iter.Next(ctx) {
		dev := DeviceWithTenantID{}
		err := iter.Decode(&dev)
		if err != nil {
			return err
		}
		if len(deviceBatch) == cap(deviceBatch) ||
			(tenantID != dev.TenantID && len(deviceBatch) > 0) {
			err := a.syncBatch(tCtx, deviceBatch, integCache, failEarly)
			if err != nil {
				return err
			}
			deviceBatch = deviceBatch[:0]
		}
		if tenantID != dev.TenantID {
			tenantID = dev.TenantID
			tCtx = identity.WithContext(ctx, &identity.Identity{
				Tenant: tenantID,
			})

			integCache, err = a.syncCacheIntegrations(tCtx)
			if err != nil {
				return err
			}

		}
		deviceBatch = append(deviceBatch, dev.Device)
	}
	if len(deviceBatch) > 0 {
		err := a.syncBatch(tCtx, deviceBatch, integCache, failEarly)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *app) DecommissionDevice(ctx context.Context, deviceID string) error {
	integrations, err := a.GetIntegrations(ctx)
	if err != nil {
		return err
	}
	var (
		errStack model.ErrorStack
		device   = newDevice(deviceID, a.store)
	)
	event := model.Event{
		WebhookEvent: model.WebhookEvent{
			ID:   uuid.New(),
			Type: model.EventTypeDeviceDecommissioned,
			Data: model.DeviceEvent{
				ID: deviceID,
			},
			EventTS: time.Now(),
		},
		DeliveryStatus: make([]model.DeliveryStatus, 0, len(integrations)),
	}
	for _, integration := range integrations {
		var (
			err error
			ok  bool
		)
		deliver := model.DeliveryStatus{
			IntegrationID: integration.ID,
			Success:       true,
		}
		switch integration.Provider {
		case model.ProviderIoTHub:
			ok, err = device.HasIntegration(ctx, integration.ID)
			if err != nil {
				break // switch
			} else if !ok {
				continue // loop
			}
			err = a.decommissionIoTHubDevice(ctx, deviceID, integration)
		case model.ProviderIoTCore:
			ok, err = device.HasIntegration(ctx, integration.ID)
			if err != nil {
				break // switch
			} else if !ok {
				continue // loop
			}
			err = a.decommissionIoTCoreDevice(ctx, deviceID, integration)
		case model.ProviderWebhook:
			var (
				req *http.Request
				rsp *http.Response
			)
			req, err = client.NewWebhookRequest(ctx,
				&integration.Credentials,
				event.WebhookEvent)
			if err != nil {
				break // switch
			}
			rsp, err = a.httpClient.Do(req)
			if err != nil {
				break // switch
			}
			deliver.StatusCode = &rsp.StatusCode
			if rsp.StatusCode >= 300 {
				err = client.NewHTTPError(rsp.StatusCode)
			}
			_ = rsp.Body.Close()

		default:
			continue
		}
		if err != nil {
			var httpError client.HTTPError
			if errors.As(err, &httpError) {
				errCode := httpError.Code()
				deliver.StatusCode = &errCode
			}
			deliver.Success = false
			deliver.Error = err.Error()
			_ = errStack.Push(err)
		}
		event.DeliveryStatus = append(event.DeliveryStatus, deliver)
	}
	err = a.store.DeleteDevice(ctx, deviceID)
	if err == store.ErrObjectNotFound {
		err = ErrDeviceNotFound
	}
	errSave := a.store.SaveEvent(ctx, event)
	if errSave != nil {
		_ = errStack.Push(errSave)
	}
	if errStack != nil {
		if err != nil {
			err = errors.Wrap(err, errStack.Error())
		} else {
			err = errStack
		}
	}
	return err
}

func (a *app) GetDevice(ctx context.Context, deviceID string) (*model.Device, error) {
	device, err := a.store.GetDevice(ctx, deviceID)
	if err == store.ErrObjectNotFound {
		return nil, ErrDeviceNotFound
	}
	return device, err
}

func (a *app) GetDeviceStateIntegration(
	ctx context.Context,
	deviceID string,
	integrationID uuid.UUID,
) (*model.DeviceState, error) {
	_, err := a.store.GetDeviceByIntegrationID(ctx, deviceID, integrationID)
	if err != nil {
		if err == store.ErrObjectNotFound {
			return nil, ErrIntegrationNotFound
		}
		return nil, errors.Wrap(err, "failed to retrieve the device")
	}
	integration, err := a.store.GetIntegrationById(ctx, integrationID)
	if integration == nil && (err == nil || err == store.ErrObjectNotFound) {
		return nil, ErrIntegrationNotFound
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the integration")
	}
	switch integration.Provider {
	case model.ProviderIoTHub:
		return a.GetDeviceStateIoTHub(ctx, deviceID, integration)
	case model.ProviderIoTCore:
		return a.GetDeviceStateIoTCore(ctx, deviceID, integration)
	default:
		return nil, ErrUnknownIntegration
	}
}

func (a *app) SetDeviceStateIntegration(
	ctx context.Context,
	deviceID string,
	integrationID uuid.UUID,
	state *model.DeviceState,
) (*model.DeviceState, error) {
	device, err := a.store.GetDeviceByIntegrationID(ctx, deviceID, integrationID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the device")
	} else if device == nil {
		return nil, ErrIntegrationNotFound
	}
	integration, err := a.store.GetIntegrationById(ctx, integrationID)
	if integration == nil && (err == nil || err == store.ErrObjectNotFound) {
		return nil, ErrIntegrationNotFound
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the integration")
	}
	switch integration.Provider {
	case model.ProviderIoTHub:
		return a.SetDeviceStateIoTHub(ctx, deviceID, integration, state)
	case model.ProviderIoTCore:
		return a.SetDeviceStateIoTCore(ctx, deviceID, integration, state)
	default:
		return nil, ErrUnknownIntegration
	}
}

func (a *app) GetEvents(ctx context.Context, filter model.EventsFilter) ([]model.Event, error) {
	return a.store.GetEvents(ctx, filter)
}

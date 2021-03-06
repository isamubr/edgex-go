package http

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"

	"github.com/gorilla/mux"
)

type DeviceProfileController struct {
	reader io.DeviceProfileReader
	dic    *di.Container
}

// NewDeviceProfileController creates and initializes an DeviceProfileController
func NewDeviceProfileController(dic *di.Container) *DeviceProfileController {
	return &DeviceProfileController{
		reader: io.NewDeviceProfileRequestReader(),
		dic:    dic,
	}
}

func (dc *DeviceProfileController) AddDeviceProfile(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	addDeviceProfileDTOs, err := dc.reader.ReadDeviceProfileRequest(r.Body)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response := commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		utils.WriteHttpHeader(w, ctx, err.Code())
		pkg.Encode(response, w, lc)
		return
	}
	deviceProfiles := requestDTO.DeviceProfileReqToDeviceProfileModels(addDeviceProfileDTOs)

	var addResponses []interface{}
	for i, d := range deviceProfiles {
		var addDeviceProfileResponse interface{}
		// get the requestID from AddDeviceProfileDTO
		reqId := addDeviceProfileDTOs[i].RequestId
		newId, err := application.AddDeviceProfile(d, ctx, dc.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			addDeviceProfileResponse = commonDTO.NewBaseResponse(
				reqId,
				err.Message(),
				err.Code())
		} else {
			addDeviceProfileResponse = commonDTO.NewBaseWithIdResponse(
				reqId,
				"Add device profiles successfully",
				http.StatusCreated,
				newId)
		}
		addResponses = append(addResponses, addDeviceProfileResponse)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.Encode(addResponses, w, lc)
}

func (dc *DeviceProfileController) UpdateDeviceProfile(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	updateDeviceProfileReq, err := dc.reader.ReadDeviceProfileRequest(r.Body)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response := commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		utils.WriteHttpHeader(w, ctx, err.Code())
		pkg.Encode(response, w, lc)
		return
	}
	deviceProfiles := requestDTO.DeviceProfileReqToDeviceProfileModels(updateDeviceProfileReq)

	var responses []interface{}
	for i, d := range deviceProfiles {
		var response interface{}
		reqId := updateDeviceProfileReq[i].RequestId
		err := application.UpdateDeviceProfile(d, ctx, dc.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(
				reqId,
				err.Message(),
				err.Code())
		} else {
			response = commonDTO.NewBaseResponse(
				reqId,
				"Update device profile successfully",
				http.StatusOK)
		}
		responses = append(responses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.Encode(responses, w, lc)
}

func (dc *DeviceProfileController) AddDeviceProfileByYaml(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	var addDeviceProfileResponse interface{}
	var statusCode int

	deviceProfileDTO, err := dc.reader.ReadDeviceProfileYaml(r)
	if err != nil {
		addDeviceProfileResponse = commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		utils.WriteHttpHeader(w, ctx, err.Code())
		pkg.Encode(addDeviceProfileResponse, w, lc)
		return
	}
	deviceProfile := dtos.ToDeviceProfileModel(deviceProfileDTO)

	newId, err := application.AddDeviceProfile(deviceProfile, ctx, dc.dic)
	if err != nil {
		addDeviceProfileResponse = commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		statusCode = err.Code()
	} else {
		addDeviceProfileResponse = commonDTO.NewBaseWithIdResponse(
			"",
			"Add device profiles successfully",
			http.StatusCreated,
			newId)
		statusCode = http.StatusCreated
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	// Encode and send the resp body as JSON format
	pkg.Encode(addDeviceProfileResponse, w, lc)
}

func (dc *DeviceProfileController) UpdateDeviceProfileByYaml(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	var response interface{}
	var statusCode int

	deviceProfileDTO, err := dc.reader.ReadDeviceProfileYaml(r)
	if err != nil {
		response = commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		utils.WriteHttpHeader(w, ctx, err.Code())
		pkg.Encode(response, w, lc)
		return
	}

	deviceProfile := dtos.ToDeviceProfileModel(deviceProfileDTO)
	err = application.UpdateDeviceProfile(deviceProfile, ctx, dc.dic)
	if err != nil {
		response = commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		statusCode = err.Code()
	} else {
		response = commonDTO.NewBaseResponse(
			"",
			"Update device profile successfully",
			http.StatusOK)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (dc *DeviceProfileController) GetDeviceProfileByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	name := vars[contractsV2.Name]

	var response interface{}
	var statusCode int

	deviceProfile, err := application.GetDeviceProfileByName(name, ctx, dc.dic)
	if err != nil {
		if errors.Kind(err) != errors.KindEntityDoesNotExist {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		}
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = responseDTO.NewDeviceProfileResponse("", "", http.StatusOK, deviceProfile)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc) // encode and send out the response
}

func (dc *DeviceProfileController) DeleteDeviceProfileById(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	id := vars[contractsV2.Id]

	var response interface{}
	var statusCode int

	err := application.DeleteDeviceProfileById(id, ctx, dc.dic)
	if err != nil {
		if errors.Kind(err) != errors.KindEntityDoesNotExist {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		}
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = commonDTO.NewBaseResponse(
			"",
			"Delete device profile successfully",
			http.StatusOK)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (dc *DeviceProfileController) DeleteDeviceProfileByName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(dc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	name := vars[contractsV2.Name]

	var response interface{}
	var statusCode int

	err := application.DeleteDeviceProfileByName(name, ctx, dc.dic)
	if err != nil {
		if errors.Kind(err) != errors.KindEntityDoesNotExist {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		}
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = commonDTO.NewBaseResponse(
			"",
			"Delete device profile successfully",
			http.StatusOK)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

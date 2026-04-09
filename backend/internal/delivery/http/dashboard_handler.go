// Package delivery chứa HTTP handler adapter – chuyển đổi HTTP request/response
// sang các lời gọi usecase. Layer này biết về HTTP nhưng không biết về DB.
package delivery

import (
	"encoding/json"
	"net/http"
	"time"

	"fusion/internal/repository"
	"fusion/internal/usecase"
)

// DashboardHandler nhóm tất cả các HTTP handler liên quan tới Dashboard.
type DashboardHandler struct {
	uc     *usecase.DashboardUsecase
	entity repository.EntityRepository
}

// NewDashboardHandler khởi tạo handler với usecase được inject vào.
func NewDashboardHandler(uc *usecase.DashboardUsecase, er repository.EntityRepository) *DashboardHandler {
	return &DashboardHandler{uc: uc, entity: er}
}

// HandleHealthz là health check endpoint đơn giản.
func HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// HandleMonthlyProduction trả về dữ liệu sản lượng theo tháng.
// Query param: ?month=YYYY-MM (mặc định: tháng hiện tại)
func (h *DashboardHandler) HandleMonthlyProduction(w http.ResponseWriter, r *http.Request) {
	var selectedMonth time.Time
	if m := r.URL.Query().Get("month"); m != "" {
		if t, err := time.Parse("2006-01", m); err == nil {
			selectedMonth = t
		}
	}

	var data interface{}
	if selectedMonth.IsZero() {
		data = h.uc.GetMonthlyProduction()
	} else {
		data = h.uc.GetMonthlyProduction(selectedMonth)
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// HandleInverterDCPower trả về dữ liệu DC/AC của một inverter.
// Query param: ?device=<device_id>
func (h *DashboardHandler) HandleInverterDCPower(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device")
	if deviceID == "" {
		http.Error(w, "Missing 'device' query parameter", http.StatusBadRequest)
		return
	}

	data := h.uc.GetInverterPower(deviceID)

	type PowerResponse struct {
		Device string      `json:"device"`
		Data   interface{} `json:"data"`
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PowerResponse{Device: deviceID, Data: data})
}

// HandleRename xử lý yêu cầu đổi tên và cấu hình chuỗi PV từ giao diện.
func (h *DashboardHandler) HandleRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Type            string `json:"type"` // "site", "logger", "device"
		ID              string `json:"id"`
		NewName         string `json:"newName"`
		StringSet       string `json:"stringSet"`       // Optional: only for devices
		ExcludedStrings string `json:"excludedStrings"` // Optional: comma-separated, only for devices
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.entity.UpdateNameChange(req.Type, req.ID, req.NewName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if req.Type == "device" && req.StringSet != "" {
		if err := h.entity.UpdateDeviceStringSet(req.ID, req.StringSet); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if req.Type == "device" {
		if err := h.entity.UpdateDeviceExcludedStrings(req.ID, req.ExcludedStrings); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

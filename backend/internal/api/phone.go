package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"provisioning-system/internal/models"
	"provisioning-system/internal/provisioner"

	"gorm.io/gorm"
)

type PhoneHandler struct {
	DB          *gorm.DB
	ProvManager *provisioner.Manager
}

func NewPhoneHandler(db *gorm.DB, pm *provisioner.Manager) *PhoneHandler {
	return &PhoneHandler{
		DB:          db,
		ProvManager: pm,
	}
}

// CreatePhone handles POST /api/phones
func (h *PhoneHandler) CreatePhone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var phone models.Phone
	if err := json.NewDecoder(r.Body).Decode(&phone); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find model in manager
	var model *provisioner.DeviceModel
	if phone.ModelID != "" {
		for _, m := range h.ProvManager.Models {
			if m.ID == phone.ModelID {
				model = &m
				break
			}
		}
	}

	// Validation Logic
	isGateway := model != nil && model.Type == "gateway"

	if !isGateway {
		if phone.MacAddress == nil || *phone.MacAddress == "" {
			http.Error(w, "MAC Address is required for phones", http.StatusBadRequest)
			return
		}
		if phone.PhoneNumber == nil || *phone.PhoneNumber == "" {
			http.Error(w, "Phone Number is required for phones", http.StatusBadRequest)
			return
		}
	} else {
		// For gateways, if MAC/Number is empty string, set to nil to allow multiple NULLs in DB
		if phone.MacAddress != nil && *phone.MacAddress == "" {
			phone.MacAddress = nil
		}
		if phone.PhoneNumber != nil && *phone.PhoneNumber == "" {
			phone.PhoneNumber = nil
		}
	}

	if model != nil {
		// Check MaxAdditionalLines
		if len(phone.Lines) > model.MaxAdditionalLines {
			http.Error(w, fmt.Sprintf("Too many additional lines. Max allowed: %d", model.MaxAdditionalLines), http.StatusBadRequest)
			return
		}
	}

	// Check for duplicate MAC
	if phone.MacAddress != nil && *phone.MacAddress != "" {
		var count int64
		h.DB.Model(&models.Phone{}).Where("mac_address = ?", *phone.MacAddress).Count(&count)
		if count > 0 {
			http.Error(w, "Phone with this MAC address already exists", http.StatusConflict)
			return
		}
	}

	// Check for duplicate Number (in Phones and PhoneLines)
	// 1. Check main phone number
	if phone.PhoneNumber != nil && *phone.PhoneNumber != "" {
		var count int64
		h.DB.Model(&models.Phone{}).Where("phone_number = ?", *phone.PhoneNumber).Count(&count)
		if count > 0 {
			http.Error(w, "Phone number already exists", http.StatusConflict)
			return
		}
		h.DB.Model(&models.PhoneLine{}).Where("phone_number = ?", *phone.PhoneNumber).Count(&count)
		if count > 0 {
			http.Error(w, "Phone number already exists in lines", http.StatusConflict)
			return
		}
	}

	// 2. Check lines numbers
	for _, line := range phone.Lines {
		var count int64
		h.DB.Model(&models.Phone{}).Where("phone_number = ?", line.PhoneNumber).Count(&count)
		if count > 0 {
			http.Error(w, fmt.Sprintf("Line number %s already exists", line.PhoneNumber), http.StatusConflict)
			return
		}
		h.DB.Model(&models.PhoneLine{}).Where("phone_number = ?", line.PhoneNumber).Count(&count)
		if count > 0 {
			http.Error(w, fmt.Sprintf("Line number %s already exists", line.PhoneNumber), http.StatusConflict)
			return
		}
	}

	if result := h.DB.Create(&phone); result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(phone)
}

// GetPhones handles GET /api/phones
// Query params: domain, vendor, mac, number, caller_id, page, limit
func (h *PhoneHandler) GetPhones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := h.DB.Model(&models.Phone{})

	// Filters
	if domain := r.URL.Query().Get("domain"); domain != "" {
		query = query.Where("domain = ?", domain)
	}
	if vendor := r.URL.Query().Get("vendor"); vendor != "" {
		query = query.Where("vendor = ?", vendor)
	}
	if mac := r.URL.Query().Get("mac"); mac != "" {
		mac = strings.ReplaceAll(mac, "*", "%")
		query = query.Where("mac_address LIKE ?", mac)
	}
	if number := r.URL.Query().Get("number"); number != "" {
		number = strings.ReplaceAll(number, "*", "%")
		query = query.Where("phone_number LIKE ?", number)
	}
	if callerID := r.URL.Query().Get("caller_id"); callerID != "" {
		callerID = strings.ReplaceAll(callerID, "*", "%")
		query = query.Where("caller_id LIKE ?", callerID)
	}

	// Count total
	var total int64
	h.DB.Model(&models.Phone{}).Where(query).Count(&total)

	// Pagination
	page := 1
	limit := 20
	if p := r.URL.Query().Get("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	offset := (page - 1) * limit

	var phones []models.Phone
	if result := query.Limit(limit).Offset(offset).Order("id desc").Preload("Lines").Find(&phones); result.Error != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch phones: %v", result.Error), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"phones": phones,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PhoneHandler) UpdatePhone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var reqPhone models.Phone
	if err := json.NewDecoder(r.Body).Decode(&reqPhone); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var existingPhone models.Phone
	if err := h.DB.Preload("Lines").First(&existingPhone, id).Error; err != nil {
		http.Error(w, "Phone not found", http.StatusNotFound)
		return
	}

	// Find model in manager
	var model *provisioner.DeviceModel
	if reqPhone.ModelID != "" {
		for _, m := range h.ProvManager.Models {
			if m.ID == reqPhone.ModelID {
				model = &m
				break
			}
		}
	}

	// Validation Logic
	isGateway := model != nil && model.Type == "gateway"

	if !isGateway {
		if reqPhone.MacAddress == nil || *reqPhone.MacAddress == "" {
			http.Error(w, "MAC Address is required for phones", http.StatusBadRequest)
			return
		}
		if reqPhone.PhoneNumber == nil || *reqPhone.PhoneNumber == "" {
			http.Error(w, "Phone Number is required for phones", http.StatusBadRequest)
			return
		}
	} else {
		// For gateways, if MAC/Number is empty string, set to nil to allow multiple NULLs in DB
		if reqPhone.MacAddress != nil && *reqPhone.MacAddress == "" {
			reqPhone.MacAddress = nil
		}
		if reqPhone.PhoneNumber != nil && *reqPhone.PhoneNumber == "" {
			reqPhone.PhoneNumber = nil
		}
	}

	if model != nil {
		if len(reqPhone.Lines) > model.MaxAdditionalLines {
			http.Error(w, fmt.Sprintf("Too many additional lines. Max allowed: %d", model.MaxAdditionalLines), http.StatusBadRequest)
			return
		}
	}

	// Check for duplicate MAC (exclude current phone)
	if reqPhone.MacAddress != nil && *reqPhone.MacAddress != "" {
		var count int64
		h.DB.Model(&models.Phone{}).Where("mac_address = ? AND id != ?", *reqPhone.MacAddress, id).Count(&count)
		if count > 0 {
			http.Error(w, "Phone with this MAC address already exists", http.StatusConflict)
			return
		}
	}

	// Check for duplicate Number (exclude current phone and its lines)
	// 1. Check main phone number
	if reqPhone.PhoneNumber != nil && *reqPhone.PhoneNumber != "" {
		var count int64
		h.DB.Model(&models.Phone{}).Where("phone_number = ? AND id != ?", *reqPhone.PhoneNumber, id).Count(&count)
		if count > 0 {
			http.Error(w, "Phone number already exists", http.StatusConflict)
			return
		}
		h.DB.Model(&models.PhoneLine{}).Where("phone_number = ? AND phone_id != ?", *reqPhone.PhoneNumber, id).Count(&count)
		if count > 0 {
			http.Error(w, "Phone number already exists in lines", http.StatusConflict)
			return
		}
	}

	// 2. Check lines numbers
	for _, line := range reqPhone.Lines {
		var count int64
		h.DB.Model(&models.Phone{}).Where("phone_number = ? AND id != ?", line.PhoneNumber, id).Count(&count)
		if count > 0 {
			http.Error(w, fmt.Sprintf("Line number %s already exists", line.PhoneNumber), http.StatusConflict)
			return
		}
		// For lines, we need to be careful. If we are updating an existing line, we should exclude it.
		// But reqPhone.Lines might contain new lines (ID=0) or existing lines (ID!=0).
		query := h.DB.Model(&models.PhoneLine{}).Where("phone_number = ?", line.PhoneNumber)
		if line.ID != 0 {
			query = query.Where("id != ?", line.ID)
		}
		query.Count(&count)
		if count > 0 {
			http.Error(w, fmt.Sprintf("Line number %s already exists", line.PhoneNumber), http.StatusConflict)
			return
		}
	}

	// Update fields
	existingPhone.Domain = reqPhone.Domain
	existingPhone.Vendor = reqPhone.Vendor
	existingPhone.ModelID = reqPhone.ModelID
	existingPhone.MacAddress = reqPhone.MacAddress
	existingPhone.PhoneNumber = reqPhone.PhoneNumber
	existingPhone.IPAddress = reqPhone.IPAddress
	existingPhone.CallerID = reqPhone.CallerID
	existingPhone.AccountSettings = reqPhone.AccountSettings
	existingPhone.Description = reqPhone.Description
	existingPhone.ExpansionModulesCount = reqPhone.ExpansionModulesCount
	existingPhone.ExpansionModuleModel = reqPhone.ExpansionModuleModel

	// Update Lines using Association Replace
	// This will delete missing lines and insert/update provided ones
	if err := h.DB.Session(&gorm.Session{FullSaveAssociations: true}).Model(&existingPhone).Association("Lines").Replace(reqPhone.Lines); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update lines: %v", err), http.StatusInternalServerError)
		return
	}

	// Save the phone itself
	if err := h.DB.Save(&existingPhone).Error; err != nil {
		http.Error(w, fmt.Sprintf("Failed to update phone: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingPhone)
}

// GetVendors handles GET /api/vendors
func (h *PhoneHandler) GetVendors(w http.ResponseWriter, r *http.Request) {
	vendors := []map[string]string{}
	for _, v := range h.ProvManager.Vendors {
		vendors = append(vendors, map[string]string{
			"id":   v.ID,
			"name": v.Name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vendors": vendors,
	})
}

// GetModels handles GET /api/models
// Query params: vendor
func (h *PhoneHandler) GetModels(w http.ResponseWriter, r *http.Request) {
	vendor := r.URL.Query().Get("vendor")
	var models []provisioner.DeviceModel

	for _, m := range h.ProvManager.Models {
		if vendor == "" || m.Vendor == vendor {
			models = append(models, m)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"models": models,
	})
}

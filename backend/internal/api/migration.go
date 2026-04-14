package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"provisioning-system/internal/logger"
	"provisioning-system/internal/models"
	"strings"

	"gorm.io/gorm"
)

type MigrationRecord struct {
	PhoneData map[string]string `json:"data"`
}

type MigrationRequest struct {
	Vendor          string              `json:"vendor"`
	Data            []map[string]string `json:"data"`
	Recommendations []interface{}       `json:"recommendations"`
}

type MigrationHandler struct {
	DB *gorm.DB
}

func NewMigrationHandler(db *gorm.DB) *MigrationHandler {
	return &MigrationHandler{DB: db}
}

/*
func (h *MigrationHandler) ApplyMigration(w http.ResponseWriter, r *http.Request) {
	var req MigrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For now, let's just implement the phone/line creation logic
	// We'll map the flattened keys to our models
	
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		for _, record := range req.Data {
			mac := record["phone.mac_address"]
			if mac == "" {
				// Try to guess from filename if mac_address mapping was not provided
				if filename, ok := record["phone.raw_filename"]; ok {
					// Very basic heuristic: extract hex from filename
					mac = strings.TrimSuffix(filename, ".cfg")
					mac = strings.TrimSuffix(mac, ".xml")
					mac = strings.TrimPrefix(mac, "SEP")
					mac = strings.TrimPrefix(mac, "spa")
				}
			}

			if mac == "" {
				continue
			}

			phoneNum := record["phone.phone_number"]
			desc := record["phone.description"]

			// Create or update phone
			var phone models.Phone
			if err := tx.Where("mac_address = ?", mac).First(&phone).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					phone = models.Phone{
						MacAddress:  &mac,
						Vendor:      req.Vendor,
						Type:        "phone",
						PhoneNumber: &phoneNum,
						Description: desc,
					}
					if err := tx.Create(&phone).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}

			// Handle Lines (generalized)
			for i := 0; i < 6; i++ {
				prefix := fmt.Sprintf("lines[%d]", i)
				lineUser := record[prefix+".user_name"]
				linePass := record[prefix+".password"]
				lineLabel := record[prefix+".label"]
				lineAccStr := record[prefix+".account_number"]

				if lineUser != "" || linePass != "" || lineLabel != "" {
					addInfo := map[string]interface{}{
						"user_name": lineUser,
						"password":  linePass,
						"label":     lineLabel,
					}
					addInfoJSON, _ := json.Marshal(addInfo)

					var line models.PhoneLine
					keyNum := i + 1
					accNum := 1
					if lineAccStr != "" {
						fmt.Sscanf(lineAccStr, "%d", &accNum)
					}

					// Check if this line exists
					if err := tx.Where("phone_id = ? AND key_number = ? AND panel_number = ?", phone.ID, keyNum, 0).First(&line).Error; err != nil {
						if err == gorm.ErrRecordNotFound {
							line = models.PhoneLine{
								PhoneID:        phone.ID,
								Type:           "Line",
								KeyNumber:      &keyNum,
								AccountNumber:  accNum,
								AdditionalInfo: string(addInfoJSON),
							}
							if err := tx.Create(&line).Error; err != nil {
								return err
							}
						} else {
							return err
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Failed to apply migration: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
*/
func (h *MigrationHandler) ApplyMigration(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain     string            `json:"domain"`
		Vendor     string            `json:"vendor"`
		ModelID    string            `json:"model_id"`
		Data       map[string]string `json:"data"`
		GlobalData map[string]string `json:"global_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Validate Uniqueness (No Overwrite Strategy)
	mac := req.Data["phone.mac_address"]
	phoneNum := req.Data["phone.phone_number"]

	if mac != "" {
		var count int64
		h.DB.Model(&models.Phone{}).Where("mac_address = ?", mac).Count(&count)
		if count > 0 {
			http.Error(w, "Conflict: MAC address already exists in system", http.StatusConflict)
			return
		}
	}

	if phoneNum != "" {
		var count int64
		h.DB.Model(&models.Phone{}).Where("phone_number = ?", phoneNum).Count(&count)
		if count > 0 {
			http.Error(w, "Conflict: Phone Number already exists in system", http.StatusConflict)
			return
		}
	}

	// 2. Map Discovery Data to System Models
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		phone := models.Phone{
			Domain:      req.Domain,
			Vendor:      req.Vendor,
			ModelID:     req.ModelID,
			MacAddress:  &mac,
			PhoneNumber: &phoneNum,
			Description: req.Data["phone.description"],
			Type:        "phone",
		}

		if err := tx.Create(&phone).Error; err != nil {
			return err
		}

		// Prepare a set of all additional_info keys from GlobalData
		globalInfo := make(map[string]interface{})
		for k, v := range req.GlobalData {
			if strings.HasPrefix(k, "feature.") {
				// feature.global_dnd.enabled -> global_dnd.enabled
				key := strings.TrimPrefix(k, "feature.")
				globalInfo[key] = v
			}
		}

		// Collect Button and Line data
		lines := make(map[int]*models.PhoneLine)
		buttons := make(map[string]*models.PhoneLine) // key: "panel-key"

		// Parse features and buttons from the specific device data
		for k, v := range req.Data {
			if strings.HasPrefix(k, "lines[") {
				// Format: lines[0].user_name
				var idx int
				fmt.Sscanf(k, "lines[%d]", &idx)
				if lines[idx] == nil {
					keyNum := idx + 1
					lines[idx] = &models.PhoneLine{
						PhoneID:       phone.ID,
						Type:          "Line",
						AccountNumber: idx + 1,
						KeyNumber:     &keyNum,
					}
				}
				
				info := make(map[string]interface{})
				if lines[idx].AdditionalInfo != "" {
					json.Unmarshal([]byte(lines[idx].AdditionalInfo), &info)
				}
				
				// Extract sub-field name (e.g., "user_name")
				field := k[strings.LastIndex(k, ".")+1:]
				info[field] = v
				
				infoJSON, _ := json.Marshal(info)
				lines[idx].AdditionalInfo = string(infoJSON)

			} else if strings.HasPrefix(k, "button.") {
				var panel, key int
				var featureId, field string
				
				if strings.HasPrefix(k, "button.ext.") {
					// Format: button.ext.1.1.speed_dial.value
					parts := strings.Split(k, ".") // [button, ext, 1, 1, speed_dial, value]
					if len(parts) < 6 { continue }
					fmt.Sscanf(parts[2], "%d", &panel)
					fmt.Sscanf(parts[3], "%d", &key)
					featureId = parts[4]
					field = parts[5]
				} else {
					// Format: button.1.speed_dial.value
					parts := strings.Split(k, ".") // [button, 1, speed_dial, value]
					if len(parts) < 4 { continue }
					fmt.Sscanf(parts[1], "%d", &key)
					panel = 0
					featureId = parts[2]
					field = parts[3]
				}
				
				mapKey := fmt.Sprintf("%d-%d", panel, key)
				if buttons[mapKey] == nil {
					pNum := panel
					kNum := key
					buttons[mapKey] = &models.PhoneLine{
						PhoneID:     phone.ID,
						Type:        featureId,
						KeyNumber:   &kNum,
						PanelNumber: &pNum,
					}
				}
				
				info := make(map[string]interface{})
				if buttons[mapKey].AdditionalInfo != "" {
					json.Unmarshal([]byte(buttons[mapKey].AdditionalInfo), &info)
				}
				
				info[field] = v
				infoJSON, _ := json.Marshal(info)
				buttons[mapKey].AdditionalInfo = string(infoJSON)
			}
		}

		// Merge Global features into all "Line" type lines and Save
		for _, line := range lines {
			info := make(map[string]interface{})
			if line.AdditionalInfo != "" {
				json.Unmarshal([]byte(line.AdditionalInfo), &info)
			}
			// Merge tags from global template if they don't exist in per-line data
			for gk, gv := range globalInfo {
				if info[gk] == nil {
					info[gk] = gv
				}
			}
			infoJSON, _ := json.Marshal(info)
			line.AdditionalInfo = string(infoJSON)
			
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}

		// Create Button lines
		for _, btn := range buttons {
			if err := tx.Create(btn).Error; err != nil {
				return err
			}
		}

		logger.Info("[Migration] Successfully processed record for MAC %s", mac)
		return nil
	})

	if err != nil {
		http.Error(w, "Migration finalize failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "mac": mac})
}

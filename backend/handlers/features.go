package handlers

import "github.com/gofiber/fiber/v2"

const (
	featureCustomers = "feature_customers_enabled"
	featureVlans     = "feature_vlans_enabled"
	featureVrfs      = "feature_vrfs_enabled"
	featureRacks     = "feature_racks_enabled"
	featureLocations = "feature_locations_enabled"
	featureBgp       = "feature_bgp_enabled"
	featureDevices   = "feature_devices_enabled"
	featureNAT       = "feature_nat_enabled"
	featureDHCP      = "feature_dhcp_enabled"
	featureCircuits  = "feature_circuits_enabled"
)

var featureResponseKeys = map[string]string{
	featureCustomers: "customers",
	featureVlans:     "vlans",
	featureVrfs:      "vrfs",
	featureRacks:     "racks",
	featureLocations: "locations",
	featureBgp:       "bgp",
	featureDevices:   "devices",
	featureNAT:       "nat",
	featureDHCP:      "dhcp",
	featureCircuits:  "circuits",
}

// GetFeatures handles GET /api/v1/features.
func (h *Handler) GetFeatures(c *fiber.Ctx) error {
	features := make(map[string]bool, len(featureResponseKeys))
	for configKey, responseKey := range featureResponseKeys {
		features[responseKey] = h.featureEnabled(c, configKey)
	}
	return c.JSON(fiber.Map{"features": features})
}

func (h *Handler) requireFeature(configKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if h.featureEnabled(c, configKey) {
			return c.Next()
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "feature disabled"})
	}
}

func (h *Handler) featureEnabled(c *fiber.Ctx, configKey string) bool {
	if h == nil || h.service == nil || h.service.Config == nil {
		return true
	}
	value, err := h.service.Config.GetCtx(c.Context(), configKey)
	if err != nil {
		return true
	}
	return value != "false"
}

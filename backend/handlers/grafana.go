package handlers

import (
	"github.com/gofiber/fiber/v2"
	"padduck/services"
)

// GrafanaHealth handles GET /api/grafana/ — required by the SimpleJSON datasource.
func (h *Handler) GrafanaHealth(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SubnetList); err != nil {
		return nil
	}
	return c.SendString("ok")
}

// grafanaSearchRequest is the payload sent by Grafana on POST /search.
type grafanaSearchRequest struct {
	Target string `json:"target"`
}

// GrafanaSearch handles POST /api/grafana/search — returns available metric names.
func (h *Handler) GrafanaSearch(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SubnetList); err != nil {
		return nil
	}
	return c.JSON([]string{
		"subnet_utilization",
		"ip_by_status",
		"section_summary",
	})
}

// grafanaQueryRequest is the payload sent by Grafana on POST /query.
type grafanaQueryRequest struct {
	Targets []struct {
		Target string `json:"target"`
		Type   string `json:"type"`
	} `json:"targets"`
}

// grafanaTableResponse is a Grafana SimpleJSON table response for one metric.
type grafanaTableResponse struct {
	Type    string          `json:"type"`
	Columns []grafanaColumn `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
}

type grafanaColumn struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

// GrafanaQuery handles POST /api/grafana/query — returns table data for requested metrics.
func (h *Handler) GrafanaQuery(c *fiber.Ctx) error {
	if err := h.permCheck(c, services.PermV2SubnetList); err != nil {
		return nil
	}

	req := new(grafanaQueryRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	var responses []grafanaTableResponse
	for _, t := range req.Targets {
		resp, err := h.buildGrafanaTable(c, t.Target)
		if err != nil {
			reqLogger(c).Error("Grafana query error", "target", t.Target, "error", err)
			continue
		}
		responses = append(responses, resp)
	}
	if responses == nil {
		responses = []grafanaTableResponse{}
	}
	return c.JSON(responses)
}

func (h *Handler) buildGrafanaTable(c *fiber.Ctx, target string) (grafanaTableResponse, error) {
	ctx := c.Context()
	switch target {
	case "subnet_utilization":
		rows, err := h.service.GrafanaSubnetUtilisation(ctx)
		if err != nil {
			return grafanaTableResponse{}, err
		}
		resp := grafanaTableResponse{
			Type: "table",
			Columns: []grafanaColumn{
				{Text: "CIDR", Type: "string"},
				{Text: "Network", Type: "string"},
				{Text: "Description", Type: "string"},
				{Text: "Used IPs", Type: "number"},
				{Text: "Total IPs", Type: "number"},
				{Text: "Utilization %", Type: "number"},
			},
		}
		for _, r := range rows {
			resp.Rows = append(resp.Rows, []interface{}{
				r.CIDR, r.NetworkName, r.Description, r.Used, r.Total, r.UtilisationPct,
			})
		}
		return resp, nil

	case "ip_by_status":
		rows, err := h.service.GrafanaIPCountsByStatus(ctx)
		if err != nil {
			return grafanaTableResponse{}, err
		}
		resp := grafanaTableResponse{
			Type: "table",
			Columns: []grafanaColumn{
				{Text: "Status", Type: "string"},
				{Text: "Count", Type: "number"},
			},
		}
		for _, r := range rows {
			resp.Rows = append(resp.Rows, []interface{}{r.Status, r.Count})
		}
		return resp, nil

	case "section_summary":
		rows, err := h.service.GrafanaNetworkSummary(ctx)
		if err != nil {
			return grafanaTableResponse{}, err
		}
		resp := grafanaTableResponse{
			Type: "table",
			Columns: []grafanaColumn{
				{Text: "Network", Type: "string"},
				{Text: "Subnets", Type: "number"},
				{Text: "Total IPs", Type: "number"},
				{Text: "Used IPs", Type: "number"},
			},
		}
		for _, r := range rows {
			resp.Rows = append(resp.Rows, []interface{}{r.NetworkName, r.SubnetCount, r.IPCount, r.UsedIPs})
		}
		return resp, nil

	default:
		return grafanaTableResponse{Type: "table", Columns: []grafanaColumn{}, Rows: [][]interface{}{}}, nil
	}
}

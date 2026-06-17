package services

// OpsManager bundles the operational sub-services that have no back-reference
// to the root Service: Discovery, Reports, Import, Jobs, Webhooks, and Topology.
// It is exposed via Service.Ops and passed directly to the Handler layer.
type OpsManager struct {
	Discovery *DiscoveryService
	Reports   *ReportsService
	Import    *ImportService
	Jobs      *JobService
	Webhooks  *WebhookService
	Topology  *TopologyService
}

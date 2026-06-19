package services

// OpsManager bundles operational sub-services that have been extracted from
// the root Service struct. It is exposed via Service.Ops and passed directly
// to the Handler layer.
type OpsManager struct {
	Discovery      *DiscoveryService
	Reports        *ReportsService
	Import         *ImportService
	Jobs           *JobService
	Webhooks       *WebhookService
	Topology       *TopologyService
	DNS            *DNSService
	Automation     *AutomationService
	Telemetry      *TelemetryService
	NetworkModules *NetworkModulesService
	IPAM           *IPAMService
	Identity       *IdentityService
	Infrastructure *InfrastructureService
}

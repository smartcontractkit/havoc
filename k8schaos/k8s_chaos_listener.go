package k8schaos

type K8sChaosListener interface {
	OnChaosCreated(chaos K8sChaos)
	OnChaosCreationFailed(chaos K8sChaos, reason error)
	OnChaosStarted(chaos K8sChaos)
	OnChaosPaused(chaos K8sChaos)
	OnChaosEnded(chaos K8sChaos)         // When the chaos is finished or deleted
	OnChaosStatusUnknown(chaos K8sChaos) // When the chaos status is unknown
	OnScheduleCreated(chaos Schedule)
	OnScheduleDeleted(chaos Schedule) // When the chaos is finished or deleted
}

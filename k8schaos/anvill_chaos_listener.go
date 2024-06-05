package k8schaos

type AnvilChaosListener interface {
	OnChaosCreated(chaos AnvilChaos)
	OnChaosCreationFailed(chaos AnvilChaos, reason error)
	OnChaosStarted(chaos AnvilChaos)
	OnChaosPaused(chaos AnvilChaos)
	OnChaosEnded(chaos AnvilChaos)         // When the chaos is finished or deleted
	OnChaosStatusUnknown(chaos AnvilChaos) // When the chaos status is unknown
}

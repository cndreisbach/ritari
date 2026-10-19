# Decisions

* [Agent and model in the actor](/decisions/agent-and-model-in-the-actor.md) - Stamp an agent as <agent>/<model> when the model is known; humans never carry a model.
* [Pi extension scope](/decisions/pi-extension-scope.md) - Build only the parts of a Pi extension a skill cannot do; no auto-triggered memory writes.
* [Session-start hooks](/decisions/session-start-hooks.md) - rit install hooks writes project-scoped hooks that run the read-only rit brief at session start.
* [Skill set after routing review](/decisions/skill-set-after-routing-review.md) - Keep ritari, ritari-ingest and ritari-reflect; fold the rest into references; move session start/end out of routing.

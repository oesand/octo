package octo

import "slices"

// CleanInjections removes all inject declarations that match the selector function.
// If selector panics, registrations remain unchanged and the panic propagates.
// Do not use octo.* functions inside selector, this may cause deadlocks
func CleanInjections(container *Container, selector func(decl Declaration) bool) {
	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	removed := make(map[Declaration]bool)
	for _, inject := range container.ordered {
		if selector(inject) {
			removed[inject] = true
		}
	}

	if len(removed) == 0 {
		return
	}
	shouldRemove := func(decl Declaration) bool { return removed[decl] }
	for groupType, group := range container.injects {
		group = slices.DeleteFunc(group, shouldRemove)
		if len(group) == 0 {
			delete(container.injects, groupType)
		} else {
			container.injects[groupType] = group
		}
	}

	container.ordered = slices.DeleteFunc(container.ordered, shouldRemove)
	container.resolveCache = nil
}

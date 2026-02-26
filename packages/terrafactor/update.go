package terrafactor

// rerenderProvider re-renders the full provider into dst, overwriting existing files.
// Identical rendering pipeline to scaffoldProvider — the separation exists so
// update-specific behaviour (e.g. partial re-render) can diverge here independently.
// Returns (rendered, copied, error).
func rerenderProvider(src, dst string, data TemplateData, resources []ResourceSpec) (int, int, error) {
	return renderResources(src, dst, data, resources)
}

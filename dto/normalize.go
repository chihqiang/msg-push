package dto

// DefaultPageSize 分页查询默认每页条数。
const DefaultPageSize = 20

// DefaultTopLimit 热门榜单默认条数。
const DefaultTopLimit = 10

// normalizePagination 补齐分页参数默认值。
func normalizePagination(page, pageSize *int) {
	if *page <= 0 {
		*page = 1
	}
	if *pageSize <= 0 {
		*pageSize = DefaultPageSize
	}
}

// Normalize 补齐分页默认值。
func (r *PageRequest) Normalize() { normalizePagination(&r.Page, &r.PageSize) }

// Normalize 补齐分页默认值。
func (r *ChannelListRequest) Normalize() { r.PageRequest.Normalize() }

// Normalize 补齐分页默认值。
func (r *TemplateListRequest) Normalize() { r.PageRequest.Normalize() }

// Normalize 补齐分页默认值。
func (r *ChannelBindingListRequest) Normalize() { normalizePagination(&r.Page, &r.PageSize) }

// Normalize 补齐分页默认值。
func (r *ChannelSignatureMappingListRequest) Normalize() { normalizePagination(&r.Page, &r.PageSize) }

// Normalize 补齐分页默认值。
func (r *FailureRuleListRequest) Normalize() { normalizePagination(&r.Page, &r.PageSize) }

// Normalize 补齐分页默认值。
func (r *LogListRequest) Normalize() { normalizePagination(&r.Page, &r.PageSize) }

// Normalize 补齐分页默认值。
func (r *CallbackListRequest) Normalize() { normalizePagination(&r.Page, &r.PageSize) }

// Normalize 补齐分页默认值。
func (r *ProviderAccountListRequest) Normalize() { normalizePagination(&r.Page, &r.PageSize) }

// Normalize 补齐分页默认值。
func (r *ProviderSignatureListRequest) Normalize() { normalizePagination(&r.Page, &r.PageSize) }

// Normalize 补齐分页默认值。
func (r *ProviderTemplateListRequest) Normalize() { normalizePagination(&r.Page, &r.PageSize) }

// Normalize 补齐分页默认值。
func (r *PushTaskListRequest) Normalize() { normalizePagination(&r.Page, &r.PageSize) }

// Normalize 补齐分页默认值。
func (r *PushBatchTaskListRequest) Normalize() { normalizePagination(&r.Page, &r.PageSize) }

// Normalize 补齐分页默认值。
func (r *WebhookConfigListRequest) Normalize() { normalizePagination(&r.Page, &r.PageSize) }

// Normalize 补齐分页默认值。
func (r *WebhookLogListRequest) Normalize() { normalizePagination(&r.Page, &r.PageSize) }

// Normalize 补齐条数默认值。
func (r *TopListRequest) Normalize() {
	if r.Limit <= 0 {
		r.Limit = DefaultTopLimit
	}
}

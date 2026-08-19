package sender

// batchTaskParams 获取批量请求第 i 个任务的模板参数映射。
// 优先使用每任务独立参数（TaskParams[i]），否则回退到共用 MappedParams。
func batchTaskParams(req *BatchSendRequest, i int) map[string]string {
	if req != nil && req.TaskParams != nil && i < len(req.TaskParams) && req.TaskParams[i] != nil {
		return req.TaskParams[i]
	}
	if req != nil {
		return req.MappedParams
	}
	return nil
}

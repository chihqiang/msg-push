package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"chihqiang/msg-push/core/sender"
	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/stringx"
	"gorm.io/gorm"
)

// ProviderAccountLogic 服务商账号管理逻辑。
type ProviderAccountLogic struct {
	svc *svc.ServiceContext
}

// NewProviderAccountLogic 创建服务商账号管理逻辑。
func NewProviderAccountLogic(s *svc.ServiceContext) *ProviderAccountLogic {
	return &ProviderAccountLogic{svc: s}
}

// GetAvailableProviders 获取可用服务商列表（可按类型过滤）。
func (l *ProviderAccountLogic) GetAvailableProviders(msgType string) []*dto.AvailableProviderResponse {
	var metas []*sender.Meta
	if msgType != "" {
		metas = sender.GetByType(msgType)
	} else {
		metas = sender.GetAll()
	}
	out := make([]*dto.AvailableProviderResponse, 0, len(metas))
	for _, m := range metas {
		out = append(out, &dto.AvailableProviderResponse{
			Code:              m.Code,
			Name:              m.Name,
			Type:              m.Type,
			Description:       m.Description,
			SupportsSend:      m.SupportsSend,
			SupportsBatchSend: m.SupportsBatchSend,
			SupportsCallback:  m.SupportsCallback,
			RequiresSignature: m.RequiresSignature,
			Website:           m.Website,
			ConsoleUrl:        m.ConsoleUrl,
			SortOrder:         m.SortOrder,
			Tags:              m.Tags,
			// 前端"新建服务商账号"弹窗依赖此字段动态渲染配置表单
			ConfigFields: toConfigFieldResponses(m.ConfigFields),
		})
	}
	return out
}

// GetProviderConfigFields 获取服务商配置字段定义。
func (l *ProviderAccountLogic) GetProviderConfigFields(providerCode string) ([]*dto.ConfigFieldResponse, error) {
	meta, ok := sender.GetByCode(providerCode)
	if !ok {
		return nil, fmt.Errorf("provider %s not found", providerCode)
	}
	return toConfigFieldResponses(meta.ConfigFields), nil
}

// toConfigFieldResponses 将服务商元信息中的配置字段转换为 DTO 响应。
func toConfigFieldResponses(fields []sender.ConfigField) []*dto.ConfigFieldResponse {
	out := make([]*dto.ConfigFieldResponse, 0, len(fields))
	for _, f := range fields {
		resp := &dto.ConfigFieldResponse{
			Key:            f.Key,
			Label:          f.Label,
			Description:    f.Description,
			Type:           f.Type,
			Required:       f.Required,
			Example:        f.Example,
			Placeholder:    f.Placeholder,
			ValidationRule: f.ValidationRule,
			DefaultValue:   f.DefaultValue,
		}
		for _, o := range f.Options {
			resp.Options = append(resp.Options, struct {
				Value string `json:"value"`
				Label string `json:"label"`
			}{Value: o.Value, Label: o.Label})
		}
		out = append(out, resp)
	}
	return out
}

// Create 创建服务商账号。
func (l *ProviderAccountLogic) Create(ctx context.Context, req *dto.CreateProviderAccountRequest) (*dto.ProviderAccountResponse, error) {
	meta, ok := sender.GetByCode(req.ProviderCode)
	if !ok {
		return nil, fmt.Errorf("provider %s not found", req.ProviderCode)
	}

	// 校验必填配置字段
	if err := validateProviderConfig(meta, req.Config); err != nil {
		return nil, err
	}

	acc := &model.ProviderAccount{
		AccountCode:  "pa_" + stringx.Randn(12, stringx.RandTypeLower),
		AccountName:  req.AccountName,
		ProviderCode: req.ProviderCode,
		ProviderType: meta.Type,
		Status:       1,
		Remark:       req.Remark,
	}
	if req.Status != nil {
		acc.Status = *req.Status
	}
	if err := acc.SetConfig(req.Config); err != nil {
		return nil, err
	}
	if err := l.svc.DB.WithContext(ctx).Create(acc).Error; err != nil {
		return nil, err
	}
	return l.toResponse(ctx, acc)
}

// List 分页查询服务商账号。
func (l *ProviderAccountLogic) List(ctx context.Context, req *dto.ProviderAccountListRequest) (*dto.ProviderAccountListResponse, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.ProviderAccount{})
	if req.ProviderType != "" {
		q = q.Where("provider_type = ?", req.ProviderType)
	}
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}
	if req.Key != "" {
		like := "%" + req.Key + "%"
		q = q.Where("account_name LIKE ? OR account_code LIKE ? OR provider_code LIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var accounts []model.ProviderAccount
	if err := q.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&accounts).Error; err != nil {
		return nil, err
	}

	resp := &dto.ProviderAccountListResponse{
		List:     make([]*dto.ProviderAccountResponse, 0, len(accounts)),
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	for i := range accounts {
		r, err := l.toResponse(ctx, &accounts[i])
		if err != nil {
			return nil, err
		}
		resp.List = append(resp.List, r)
	}
	return resp, nil
}

// Get 获取服务商账号详情。
func (l *ProviderAccountLogic) Get(ctx context.Context, id uint) (*dto.ProviderAccountResponse, error) {
	var acc model.ProviderAccount
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&acc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("provider account not found")
		}
		return nil, err
	}
	return l.toResponse(ctx, &acc)
}

// Update 更新服务商账号。
func (l *ProviderAccountLogic) Update(ctx context.Context, id uint, req *dto.UpdateProviderAccountRequest) error {
	var acc model.ProviderAccount
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&acc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("provider account not found")
		}
		return err
	}

	updates := map[string]any{}
	if req.AccountName != "" {
		updates["account_name"] = req.AccountName
	}
	if req.Config != nil {
		meta, ok := sender.GetByCode(acc.ProviderCode)
		if !ok {
			return fmt.Errorf("provider %s not found", acc.ProviderCode)
		}
		if err := validateProviderConfig(meta, req.Config); err != nil {
			return err
		}
		cfgJSON, err := marshalProviderConfig(req.Config)
		if err != nil {
			return err
		}
		updates["config"] = cfgJSON
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if len(updates) == 0 {
		return nil
	}
	return l.svc.DB.WithContext(ctx).Model(&model.ProviderAccount{}).
		Where("id = ?", id).Updates(updates).Error
}

// Delete 删除服务商账号。
func (l *ProviderAccountLogic) Delete(ctx context.Context, id uint) error {
	res := l.svc.DB.WithContext(ctx).Delete(&model.ProviderAccount{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("provider account not found")
	}
	return nil
}

// toResponse 模型转响应。
func (l *ProviderAccountLogic) toResponse(ctx context.Context, acc *model.ProviderAccount) (*dto.ProviderAccountResponse, error) {
	cfg, err := acc.GetConfig()
	if err != nil {
		return nil, err
	}
	providerName := acc.ProviderCode
	if meta, ok := sender.GetByCode(acc.ProviderCode); ok {
		providerName = meta.Name
	}
	return &dto.ProviderAccountResponse{
		ID:           acc.ID,
		AccountCode:  acc.AccountCode,
		AccountName:  acc.AccountName,
		ProviderCode: acc.ProviderCode,
		ProviderName: providerName,
		ProviderType: acc.ProviderType,
		Config:       cfg,
		Status:       acc.Status,
		Remark:       acc.Remark,
		CreatedAt:    acc.CreatedAt,
		UpdatedAt:    acc.UpdatedAt,
	}, nil
}

// validateProviderConfig 校验服务商必填配置字段。
func validateProviderConfig(meta *sender.Meta, cfg map[string]any) error {
	for _, f := range meta.ConfigFields {
		if !f.Required {
			continue
		}
		if v, ok := cfg[f.Key]; !ok || fmt.Sprint(v) == "" {
			return fmt.Errorf("field %s (%s) is required", f.Key, f.Label)
		}
	}
	return nil
}

// marshalProviderConfig 序列化配置为 JSON 字符串。
func marshalProviderConfig(cfg map[string]any) (string, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

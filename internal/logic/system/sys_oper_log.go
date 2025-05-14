package system

import (
	"context"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/grpool"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"net/url"
	"sagooiot/internal/consts"
	"sagooiot/internal/dao"
	"sagooiot/internal/model"
	"sagooiot/internal/model/entity"
	"sagooiot/internal/service"
	"sagooiot/pkg/utility/utils"
	"strings"
)

type sSysOperLog struct {
}

func SysOperLog() *sSysOperLog {
	return &sSysOperLog{}
}

func init() {
	service.RegisterSysOperLog(SysOperLog())
}

// GetList 获取操作日志数据列表
func (s *sSysOperLog) GetList(ctx context.Context, input *model.SysOperLogDoInput) (total int, out []*model.SysOperLogOut, err error) {
	m := dao.SysOperLog.Ctx(ctx)
	if input.Title != "" {
		m = m.WhereLike(dao.SysOperLog.Columns().Title, "%"+input.Title+"%")
	}
	if input.BusinessType != "" {
		m = m.WhereLike(dao.SysOperLog.Columns().BusinessType, "%"+input.BusinessType+"%")
	}
	if input.Method != "" {
		m = m.WhereLike(dao.SysOperLog.Columns().Method, "%"+input.Method+"%")
	}
	if input.RequestMethod != "" {
		m = m.WhereLike(dao.SysOperLog.Columns().RequestMethod, "%"+input.RequestMethod+"%")
	}
	if input.OperatorType != "" {
		m = m.WhereLike(dao.SysOperLog.Columns().OperatorType, "%"+input.OperatorType+"%")
	}
	if input.OperName != "" {
		m = m.WhereLike(dao.SysOperLog.Columns().OperName, "%"+input.OperName+"%")
	}
	if input.DeptName != "" {
		m = m.WhereLike(dao.SysOperLog.Columns().DeptName, "%"+input.DeptName+"%")
	}
	if input.OperUrl != "" {
		m = m.WhereLike(dao.SysOperLog.Columns().OperUrl, "%"+input.OperUrl+"%")
	}
	if input.OperIp != "" {
		m = m.WhereLike(dao.SysOperLog.Columns().OperIp, "%"+input.OperIp+"%")
	}
	if input.OperLocation != "" {
		m = m.WhereLike(dao.SysOperLog.Columns().OperLocation, "%"+input.OperLocation+"%")
	}
	if input.Status != -1 {
		m = m.Where(dao.SysOperLog.Columns().Status, input.Status)
	}
	//获取总数
	total, err = m.Count()
	if err != nil {
		err = gerror.New("获取操作日志列表数据失败")
		return
	}
	if input.PageNum == 0 {
		input.PageNum = 1
	}
	if input.PageSize == 0 {
		input.PageSize = consts.PageSize
	}
	//获取操作日志列表信息
	err = m.Page(input.PageNum, input.PageSize).OrderDesc(dao.SysOperLog.Columns().OperId).Scan(&out)
	if err != nil {
		err = gerror.New("获取操作日志列表失败")
		return
	}
	return
}

func (s *sSysOperLog) Invoke(ctx context.Context, userId int, url *url.URL, param g.Map, method string, clientIp string, res map[string]interface{}, err error) {
	pool := grpool.New(100)
	if err := pool.Add(ctx, func(ctx context.Context) {
		//写入操作数据
		err = s.Add(ctx, userId, url, param, method, clientIp, res, err)
	},
	); err != nil {
		g.Log().Error(ctx, err.Error())
	}
}

// Add 添加操作日志
func (s *sSysOperLog) Add(ctx context.Context, userId int, url *url.URL, param g.Map, method string, clientIp string, res map[string]interface{}, erro error) (err error) {
	operLogInfo := &entity.SysOperLog{
		OperTime: gtime.Now(),
	}

	// 查询用户信息
	var userInfo *entity.SysUser
	err = dao.SysUser.Ctx(ctx).Where(g.Map{
		dao.SysUser.Columns().Id:        userId,
		dao.SysUser.Columns().IsDeleted: 0,
		dao.SysUser.Columns().Status:    1,
	}).Scan(&userInfo)
	if err != nil {
		return err
	}
	if userInfo != nil {
		operLogInfo.OperName = userInfo.UserName

		// 查询部门信息
		var deptInfo *entity.SysDept
		err = dao.SysDept.Ctx(ctx).Where(g.Map{
			dao.SysDept.Columns().DeptId:    userInfo.DeptId,
			dao.SysDept.Columns().IsDeleted: 0,
			dao.SysDept.Columns().Status:    1,
		}).Scan(&deptInfo)
		if err != nil {
			return err
		}
		if deptInfo != nil {
			operLogInfo.DeptName = deptInfo.DeptName
		}
	}

	// 请求地址和方法
	operLogInfo.Method = url.Path
	apiInfo, apiErr := service.SysApi().GetInfoByAddress(ctx, url.Path)
	if apiErr == nil && apiInfo != nil {
		operLogInfo.Title = apiInfo.Name
	}

	// 请求方法和业务类型
	operLogInfo.RequestMethod = method
	businessTypeMap := map[string]int{
		"POST":   1,
		"PUT":    2,
		"DELETE": 3,
	}
	if bt, ok := businessTypeMap[strings.ToUpper(method)]; ok {
		operLogInfo.BusinessType = bt
	} else {
		operLogInfo.BusinessType = 0
	}
	operLogInfo.OperatorType = 1

	// 请求URL
	operLogInfo.OperUrl = url.Path + func() string {
		if url.RawQuery != "" {
			return "?" + url.RawQuery
		}
		return ""
	}()

	// 客户端IP和位置
	operLogInfo.OperIp = clientIp
	operLogInfo.OperLocation = utils.GetCityByIp(clientIp)

	// 参数
	if param != nil {
		if b, _ := gjson.Encode(param); len(b) > 0 {
			operLogInfo.OperParam = string(b)
		}
	}

	// 错误码和状态
	var code gcode.Code
	if erro != nil {
		code = gerror.Code(erro)
		if code == gcode.CodeNil {
			code = gcode.CodeInternalError
		}
	}
	if code != nil && code.Code() != 0 {
		operLogInfo.Status = 0
		errMsg := erro.Error()
		errorMsgMap := map[string]interface{}{
			"code":    code.Code(),
			"message": errMsg,
		}
		if b, _ := gjson.Encode(errorMsgMap); len(b) > 0 {
			operLogInfo.ErrorMsg = string(b)
		}
	} else {
		operLogInfo.Status = 1
		if res != nil {
			if b, _ := gjson.Encode(res); len(b) > 0 {
				operLogInfo.JsonResult = string(b)
			}
		}
	}

	_, err = dao.SysOperLog.Ctx(ctx).Data(operLogInfo).Insert()
	return
}

func (s *sSysOperLog) AnalysisLog(ctx context.Context) (data entity.SysOperLog) {
	// 获取当前请求的上下文对象
	mctx := service.Context().Get(ctx)
	request := ghttp.RequestFromCtx(ctx)
	handlerResponse := request.GetHandlerResponse() // 响应结果
	param := request.GetMap()                       // 请求参数

	res := gconv.Map(handlerResponse)

	//takeUpTime, ok := mctx.Data["request.takeUpTime"].(int64)
	//if !ok {
	//	takeUpTime = 0 // 或适当的默认值
	//}
	//g.Log().Debug(ctx, "request.takeUpTime: ", takeUpTime)

	operLogInfo := entity.SysOperLog{}

	if user := mctx.User; user != nil {
		operLogInfo.OperName = user.UserName

		var deptInfo *entity.SysDept
		err := dao.SysDept.Ctx(ctx).Where(g.Map{
			dao.SysDept.Columns().DeptId:    user.DeptId,
			dao.SysDept.Columns().IsDeleted: 0,
			dao.SysDept.Columns().Status:    1,
		}).Scan(&deptInfo)
		if err == nil && deptInfo != nil {
			operLogInfo.DeptName = deptInfo.DeptName
		}
	}

	operLogInfo.Method = request.URL.Path
	apiInfo, _ := service.SysApi().GetInfoByAddress(ctx, request.URL.Path)
	if apiInfo != nil {
		operLogInfo.Title = apiInfo.Name
	}

	operLogInfo.RequestMethod = request.Method
	operLogInfo.OperatorType = 1

	// 业务类型
	switch request.Method {
	case "POST":
		operLogInfo.BusinessType = 1
	case "PUT":
		operLogInfo.BusinessType = 2
	case "DELETE":
		operLogInfo.BusinessType = 3
	default:
		operLogInfo.BusinessType = 0
	}

	rawQuery := ""
	if rq := request.URL.RawQuery; rq != "" {
		rawQuery = "?" + rq
	}
	operLogInfo.OperUrl = request.URL.Path + rawQuery

	operLogInfo.OperIp = utils.GetClientIp(request.GetCtx())
	operLogInfo.OperLocation = utils.GetCityByIp(operLogInfo.OperIp)

	operTime, err := gtime.StrToTimeFormat(gtime.Datetime(), "2006-01-02 15:04:05")
	if err != nil {
		g.Log().Error(ctx, "Failed to parse time: ", err)
		return
	}
	operLogInfo.OperTime = operTime

	if param != nil {
		if b, err := gjson.Encode(param); err == nil {
			operLogInfo.OperParam = string(b)
		}
	}

	var code gcode.Code = gcode.CodeOK
	if erro := request.GetError(); erro != nil {
		code = gerror.Code(erro)
		if code == gcode.CodeNil {
			code = gcode.CodeInternalError
		}
		operLogInfo.Status = 0
		errorMsgMap := map[string]interface{}{
			"code":    code.Code(),
			"message": erro.Error(),
		}
		if errorMsg, err := gjson.Encode(errorMsgMap); err == nil {
			operLogInfo.ErrorMsg = string(errorMsg)
		}
	} else {
		operLogInfo.Status = 1
		if b, err := gjson.Encode(res); err == nil {
			operLogInfo.JsonResult = string(b)
			if len(operLogInfo.JsonResult) > 65535 {
				operLogInfo.JsonResult = "数据过大，未记录"
			}
		}
	}

	return operLogInfo
}

// RealWrite 真实写入
func (s *sSysOperLog) RealWrite(ctx context.Context, log entity.SysOperLog) (err error) {
	_, err = dao.SysOperLog.Ctx(ctx).FieldsEx(dao.SysOperLog.Columns().OperId).Data(log).Insert()
	return
}

// Detail 操作日志详情
func (s *sSysOperLog) Detail(ctx context.Context, operId int) (entity *entity.SysOperLog, err error) {
	_ = dao.SysOperLog.Ctx(ctx).Where(g.Map{
		dao.SysOperLog.Columns().OperId: operId,
	}).Scan(&entity)
	if entity == nil {
		return nil, gerror.New("日志ID错误")
	}
	return
}

// Del 根据ID删除操作日志
func (s *sSysOperLog) Del(ctx context.Context, operIds []int) (err error) {
	for _, operId := range operIds {
		var sysOperLog *entity.SysOperLog
		_ = dao.SysOperLog.Ctx(ctx).Where(g.Map{
			dao.SysOperLog.Columns().OperId: operId,
		}).Scan(&sysOperLog)
		if sysOperLog == nil {
			return gerror.New("ID错误")
		}
	}
	//删除操作日志
	_, err = dao.SysOperLog.Ctx(ctx).WhereIn(dao.SysOperLog.Columns().OperId, operIds).Delete()
	return
}

func (s *sSysOperLog) ClearOperationLogByDays(ctx context.Context, days int) (err error) {
	_, err = dao.SysOperLog.Ctx(ctx).Delete("to_days(now())-to_days(`oper_time`) > ?", days+1)
	return
}

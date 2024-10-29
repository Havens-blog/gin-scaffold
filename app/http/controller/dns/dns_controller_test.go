package dns

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"goskeleton/app/global/consts"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockSRegion 是 SRegion 的 mock 实现
type MockSRegion struct {
	mock.Mock
}

// DescribeSubDomainRecords 是 DescribeSubDomainRecords 的 mock 实现
func (m *MockSRegion) DescribeSubDomainRecords(domainName string, pageNumber int, pageSize int) (cloudprovider.SDomainRecordSet, error) {
	args := m.Called(domainName, pageNumber, pageSize)
	return args.Get(0).(cloudprovider.SDomainRecordSet), args.Error(1)
}

// AddDomainRecord 是 AddDomainRecord 的 mock 实现
func (m *MockSRegion) AddDomainRecord(domainName string, opts cloudprovider.DnsRecordSet) (string, error) {
	args := m.Called(domainName, opts)
	return args.String(0), args.Error(1)
}

func TestDnsRegister_Register(t *testing.T) {
	// 初始化 Gin 引擎
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// 创建 MockSRegion 实例
	mockRegion := new(MockSRegion)

	// 设置 mock 行为
	mockRegion.On("DescribeSubDomainRecords", "example.com", 1, 20).Return(cloudprovider.SDomainRecordSet{TotalCount: 0}, nil)
	mockRegion.On("AddDomainRecord", "example.com", mock.Anything).Return("recordid123", nil)

	// 替换全局变量 region
	region = mockRegion

	// 创建 DnsRegister 实例
	dnsRegister := DnsRegister{}

	// 测试用的 HTTP 请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/dns", nil)
	req.Header.Set("Content-Type", "application/json")

	// 设置上下文，并把请求赋值给它
	context, _ := gin.CreateTestContext(w)
	context.Request = req

	// 模拟上下文中的 GetString 方法
	context.Set(consts.ValidatorPrefix+"dns_content", "example.com 192.168.1.1 A\n")

	// 调用 Register 方法
	dnsRegister.Register(context)

	// 检查响应
	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"example.com":"域名解析添加成功 recordid:recordid123"}`, w.Body.String())

	// 确保所有 mock 调用都被正确执行
	mockRegion.AssertExpectations(t)
}

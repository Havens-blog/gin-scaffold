package dns

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"goskeleton/app/global/consts"
	"goskeleton/app/global/variable"
	"goskeleton/app/utils/response"
	"net"
	"regexp"
	"strings"
	"yunion.io/x/cloudmux/pkg/cloudprovider"
	"yunion.io/x/cloudmux/pkg/multicloud/aliyun"
)

type SDomainRecordListOptions struct {
	DOMAINNAME string
	PageNumber int `help:"page size" default:"1"`
	PageSize   int `help:"page PageSize" default:"20"`
}

var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63}$`)

const (
	DnsAlreadyExist  string = "域名解析已存在"
	DnsNotExist      string = "域名解析不存在"
	DnsRecordSuccess string = "域名解析添加成功"
	DnsRecordFail    string = "域名解析添加失败"
)

var region *aliyun.SRegion

func init() {
	//初始化
	options := &BaseOptions{
		Debug:     false,
		AccessKey: "LTAI5t7rVp7kVYP8CB5nd4Ym",
		Secret:    "0vDa8AebxtaWv6KcHHuI0dIBsiiuam",
		RegionId:  "cn-shenzhen",
	}
	region, _ = newClient(options)
}

type DnsRegister struct {
}

func (d *DnsRegister) Register(context *gin.Context) {
	dns_content := context.GetString(consts.ValidatorPrefix + "dns_content")
	single_record := context.GetBool(consts.ValidatorPrefix + "single_record")
	//dns_content 是一个三段式字符串结构 “域名 解析值 类型”
	lines := strings.Split(dns_content, "\n")
	myMap := make(map[string]string)
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}
		columns := strings.Fields(trimmedLine)
		if len(columns) < 2 || len(columns) > 3 {
			continue
		}

		columns = append(columns, "")
		fullDomain := columns[1]
		dnsType := columns[2]
		value := columns[0]
		if !isDomain(fullDomain) {
			myMap[fullDomain] = "域名格式不正确"
			variable.ZapLog.Sugar().Warnf("域名格式不正确: %s", fullDomain)
			continue
		}
		//添加域名解析
		subdomains, domain := ExtractSubdomains(fullDomain)
		if len(subdomains) == 0 {
			variable.ZapLog.Sugar().Warn("没有子域名")
			continue
		}

		// 将子域名组合成一个字符串
		subdomain := strings.Join(subdomains, ".")

		opt1 := SDomainRecordListOptions{
			DOMAINNAME: fullDomain,
			PageNumber: 1,
			PageSize:   20,
		}
		srecords, e := region.GetClient().DescribeSubDomainRecords(opt1.DOMAINNAME, opt1.PageNumber, opt1.PageSize)
		if e != nil {
			myMap[fullDomain] = "获取域名解析失败,检查域名是否存在当前阿里云账号"
			continue
		}
		if dnsType == "" {
			dnsType = ChargeValue(value)
			if dnsType == "" {
				myMap[fullDomain] = "解析类型不正确"
				continue
			}
		}
		args := DomainRecordCreateOptions{
			DOMAINNAME: domain,
			NAME:       subdomain,
			VALUE:      value,
			TTL:        600,
			TYPE:       dnsType,
		}
		opts := cloudprovider.DnsRecordSet{}
		opts.DnsName = args.NAME
		opts.DnsType = cloudprovider.TDnsType(args.TYPE)
		opts.DnsValue = args.VALUE
		opts.Ttl = args.TTL
		opts.PolicyType = cloudprovider.TDnsPolicyType(args.PolicyType)
		if srecords.TotalCount == 0 || single_record {

			recordId, err := region.GetClient().AddDomainRecord(args.DOMAINNAME, opts)
			if err != nil {
				myMap[fullDomain] = DnsRecordFail
				variable.ZapLog.Sugar().Error(e)
			} else {
				myMap[fullDomain] = DnsRecordSuccess + " recordid:" + recordId
			}

		} else {
			myMap[fullDomain] = DnsAlreadyExist
		}
	}
	response.Success(context, consts.CurdStatusOkMsg, myMap)
}

// 验证字符串是否为域名格式
func isDomain(domain string) bool {
	return domainRegex.MatchString(domain)
}

type BaseOptions struct {
	Debug      bool   `help:"debug mode"`
	CloudEnv   string `help:"Cloud environment" default:"$ALIYUN_CLOUD_ENV" choices:"InternationalCloud|FinanceCloud" metavar:"ALIYUN_CLOUD_ENV"`
	AccessKey  string `help:"Access key" default:"$ALIYUN_ACCESS_KEY" metavar:"ALIYUN_ACCESS_KEY"`
	Secret     string `help:"Secret" default:"$ALIYUN_SECRET" metavar:"ALIYUN_SECRET"`
	RegionId   string `help:"RegionId" default:"$ALIYUN_REGION" metavar:"ALIYUN_REGION"`
	AccountId  string `help:"AccountId" default:"$ALIYUN_ACCOUNT_ID" metavar:"ALIYUN_ACCOUNT_ID"`
	SUBCOMMAND string `help:"aliyuncli subcommand" subcommand:"true"`
}

type DomainRecordCreateOptions struct {
	DOMAINNAME  string
	NAME        string
	VALUE       string `help:"dns record value"`
	TTL         int64  `help:"ttl"`
	TYPE        string `help:"dns type"`
	PolicyType  string `help:"PolicyType"`
	PolicyValue string
}

func newClient(options *BaseOptions) (*aliyun.SRegion, error) {
	if len(options.AccessKey) == 0 {
		return nil, fmt.Errorf("Missing accessKey")
	}

	if len(options.Secret) == 0 {
		return nil, fmt.Errorf("Missing secret")
	}

	cli, err := aliyun.NewAliyunClient(
		aliyun.NewAliyunClientConfig(
			options.CloudEnv,
			options.AccessKey,
			options.Secret,
		),
	)
	if err != nil {
		return nil, err
	}

	region := cli.GetRegion(options.RegionId)
	if region == nil {
		return nil, fmt.Errorf("No such region %s", options.RegionId)
	}

	return region, nil
}

func ExtractSubdomains(fullDomain string) ([]string, string) {
	parts := strings.Split(fullDomain, ".")
	if len(parts) <= 2 {
		// 如果没有子域名，返回空切片和完整域名
		return nil, fullDomain
	}

	var subdomains []string
	for i := 0; i < len(parts)-2; i++ {
		subdomains = append(subdomains, parts[i])
	}

	domain := strings.Join(parts[len(parts)-2:], ".")
	return subdomains, domain
}

func ChargeValue(value string) string {
	// 判断是否为 A 记录
	if net.ParseIP(value) != nil {
		return "A"
	}
	ns, err := net.LookupNS(value)
	if err == nil && len(ns) > 0 {
		return "NS"
	}
	mx, err := net.LookupMX(value)
	if err == nil && len(mx) > 0 {
		return "MX"
	}
	// 判断是否为 CNAME 记录
	cname, err := net.LookupCNAME(value)
	if err == nil && cname != "" {
		return "CNAME"
	}
	// 判断是否为 TXT 记录
	txts, err := net.LookupTXT(value)
	if err == nil && len(txts) > 0 {
		return "TXT"
	}
	return ""
}
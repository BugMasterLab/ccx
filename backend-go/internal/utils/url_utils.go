package utils

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// reURLPassword 匹配 URL 中 user:password@ 的密码部分
var reURLPassword = regexp.MustCompile(`(://[^:@/]+:)[^@]+(@)`)

// RedactURLCredentials 对 URL 中的用户名和密码进行脱敏处理
// 例如: http://user:pass@host:port -> http://user:***@host:port
// 若 URL 解析失败，使用正则兜底替换，避免凭证泄露
func RedactURLCredentials(rawURL string) string {
	if rawURL == "" {
		return rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		// 解析失败时用正则兜底，避免凭证明文出现在日志中
		return reURLPassword.ReplaceAllString(rawURL, "${1}***${2}")
	}

	if u.User != nil {
		username := u.User.Username()
		// 构建脱敏后的 Userinfo
		u.User = url.UserPassword(username, "***")
		return u.String()
	}

	return rawURL
}

// ValidateBaseURL 验证 baseURL 是否安全，防止 SSRF 攻击
// 仅拦截云元数据服务（169.254.169.254），允许其他内网地址（支持 Ollama、内网部署）
func ValidateBaseURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("baseURL 不能为空")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("无效的 URL 格式: %w", err)
	}

	// 检查协议
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("不支持的协议: %s（仅允许 http/https）", u.Scheme)
	}

	// 提取主机名（去除端口）
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("URL 缺少主机名")
	}

	// 硬编码拦截云元数据服务（最关键的安全风险）
	if host == "169.254.169.254" {
		return fmt.Errorf("禁止访问云元数据服务")
	}

	// 检查域名是否解析到云元数据服务
	ips, err := net.LookupIP(host)
	if err != nil {
		// DNS 解析失败，允许通过（避免误杀）
		return nil
	}

	for _, resolvedIP := range ips {
		if resolvedIP.String() == "169.254.169.254" {
			return fmt.Errorf("域名 %s 解析到云元数据服务", host)
		}
	}

	return nil
}

// isPrivateIP 判断 IP 是否为私有地址（保留用于其他场景）
func isPrivateIP(ip net.IP) bool {
	// IPv4 私有地址段
	privateIPv4Blocks := []string{
		"10.0.0.0/8",     // Class A 私有网络
		"172.16.0.0/12",  // Class B 私有网络
		"192.168.0.0/16", // Class C 私有网络
		"127.0.0.0/8",    // Loopback
		"169.254.0.0/16", // Link-local
		"0.0.0.0/8",      // 当前网络
		"224.0.0.0/4",    // 组播
		"240.0.0.0/4",    // 保留
	}

	// IPv6 私有地址段
	privateIPv6Blocks := []string{
		"::1/128",   // Loopback
		"fc00::/7",  // Unique local
		"fe80::/10", // Link-local
		"ff00::/8",  // 组播
	}

	blocks := privateIPv4Blocks
	if ip.To4() == nil {
		blocks = append(blocks, privateIPv6Blocks...)
	}

	for _, block := range blocks {
		_, subnet, err := net.ParseCIDR(block)
		if err != nil {
			continue
		}
		if subnet.Contains(ip) {
			return true
		}
	}

	// 检查 localhost 域名
	if strings.EqualFold(ip.String(), "localhost") {
		return true
	}

	return false
}

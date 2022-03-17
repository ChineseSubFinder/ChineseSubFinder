package settings

type SubShareCenter struct {
	SenderSMTPAddress       string `json:"sender_smtp_address"`        // 发件人的 SMTP 服务地址和端口 outlook.office365.com:993
	SenderSMTPPort          int    `json:"sender_smtp_port"`           // SMTP 的端口号
	InsecureSkipVerify      bool   `json:"insecure_skip_verify"`       // 是否允许不安全连接
	SenderEmailAddress      string `json:"sender_email_address"`       // 发件人的邮件地址
	SenderEmailPwd          string `json:"sender_email_pwd"`           // 发件人的邮件地址对应的密码
	ShareCenterEmailAddress string `json:"share_center_email_address"` // 收取邮件的目标邮件地址
}

func NewSubShareCenter(SenderSMTPAddress string, SenderSMTPPort int, InsecureSkipVerify bool, SenderEmailAddress, SenderEmailPwd, ShareCenterEmailAddress string) *SubShareCenter {
	return &SubShareCenter{
		SenderSMTPAddress,
		SenderSMTPPort,
		InsecureSkipVerify,
		SenderEmailAddress,
		SenderEmailPwd,
		ShareCenterEmailAddress,
	}
}

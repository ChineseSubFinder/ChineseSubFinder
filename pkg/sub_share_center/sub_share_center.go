package sub_share_center

import (
	"crypto/tls"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/settings"
	goMail "gopkg.in/mail.v2"
)

type SubShareCenter struct {
	center settings.SubShareCenter
	client *goMail.Dialer
	dialer goMail.SendCloser
}

func NewSubShareCenter(center settings.SubShareCenter) *SubShareCenter {

	nowClient := goMail.NewDialer(center.SenderSMTPAddress, center.SenderSMTPPort, center.SenderEmailAddress, center.SenderEmailPwd)
	nowClient.TLSConfig = &tls.Config{InsecureSkipVerify: center.InsecureSkipVerify}
	return &SubShareCenter{
		center: center,
		client: nowClient,
	}
}

func (s *SubShareCenter) Login() error {

	var err error
	s.dialer, err = s.client.Dial()
	if err != nil {
		return err
	}

	return nil
}

func (s *SubShareCenter) Logout() error {
	return s.dialer.Close()
}

func (s *SubShareCenter) SendSubtitle(imdbID, videoName, feature, subFIleFPath string) error {

	return nil
}

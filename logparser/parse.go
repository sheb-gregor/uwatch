package logparser

import (
	"errors"
	"regexp"
	"time"

	"github.com/sheb-gregor/uwatch/models"
)

var (
	ErrNotSSHdLog        = errors.New("not a sshd logline or invalid")
	ErrUnSupportedStatus = errors.New("sshd action status unknown or not supported")
	ErrInvalidLine       = errors.New("not a sshd logline or invalid")
)

func ParseLine(logLine string) (*models.AuthInfo, error) {
	sshdReg, _ := regexp.Compile(`([\w+\s]+\d{2}:\d{2}:\d{2})\s(\w+)\s(sshd\[\d+\]:)\s(Accepted|Disconnected|Failed)`)
	matches := sshdReg.FindStringSubmatch(logLine)
	if len(matches) < 5 {
		return nil, ErrNotSSHdLog
	}

	timeStamp, err := time.Parse(time.Stamp, matches[1])
	if err != nil {
		return nil, err
	}

	timeStamp = timeStamp.AddDate(time.Now().Year(), 0, 0)
	authInfo := &models.AuthInfo{
		Status: models.AuthStatus(matches[4]),
		Date:   timeStamp,
	}

	switch models.AuthStatus(matches[4]) {
	case models.AuthAccepted:
		acceptedReg, _ := regexp.Compile(`(Accepted)\s(\w+)\sfor\s(\w+)\sfrom\s([\d|\.]+)`)
		matches = acceptedReg.FindStringSubmatch(logLine)
		if len(matches) < 5 {
			return nil, ErrInvalidLine
		}

		authInfo.AuthMethod = matches[2]
		authInfo.Username = matches[3]
		authInfo.RemoteAddr = matches[4]

	case models.AuthDisconnected:
		acceptedReg, _ := regexp.Compile(`(Disconnected)\sfrom user\s(\w+)\s([\d|\.]+)`)
		matches = acceptedReg.FindStringSubmatch(logLine)
		if len(matches) < 4 {
			return nil, ErrInvalidLine
		}

		authInfo.Username = matches[2]
		authInfo.RemoteAddr = matches[3]

	case models.AuthFailed:
		acceptedReg, _ := regexp.Compile(`(Failed)\s(\w+)\sfor(\sinvalid\suser)?\s(\w+)\sfrom\s([\d|\.]+)`)
		matches = acceptedReg.FindStringSubmatch(logLine)
		if len(matches) < 6 {
			return nil, ErrInvalidLine
		}

		authInfo.AuthMethod = matches[2]
		authInfo.Username = matches[4]
		authInfo.RemoteAddr = matches[5]

	default:
		return nil, ErrUnSupportedStatus
	}

	return authInfo, nil
}

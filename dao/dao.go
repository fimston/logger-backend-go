// types.go
package dao

import (
	"strconv"
	"time"
)

const TrafficIsUnlimited = -1

type AccountInfo struct {
	userId         int64
	payedTill      time.Time
	totalTraffic   float64
	allowedTraffic float64
	messageTtl     time.Duration
}

func (self *AccountInfo) IndexAlias() string {
	return strconv.FormatInt(self.userId, 10)
}

func (self *AccountInfo) Expired() bool {
	return self.payedTill.Before(time.Now())
}

func (self *AccountInfo) TrafficExceeded() bool {
	if self.allowedTraffic == TrafficIsUnlimited {
		return false
	} else {
		return self.totalTraffic >= self.allowedTraffic
	}
}

func NewAccountInfo(userId int64) *AccountInfo {
	return &AccountInfo{
		userId:         userId,
		totalTraffic:   0,
		allowedTraffic: TrafficIsUnlimited}
}

type ApiKeyMap map[string]*AccountInfo

type AccountsDao interface {
	LoadAccountsByApiKey(dest ApiKeyMap) error
}

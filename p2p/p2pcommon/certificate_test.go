package p2pcommon

import (
	"testing"
	"time"
)

func TestAgentCertificateV1_IsValidInTime(t *testing.T) {
	ct := time.Now()
	et := ct.Add(time.Hour * 24)

	type args struct {
		t            time.Time
		errTolerance time.Duration
	}
	tests := []struct {
		name string

		args args
		want bool
	}{
		{"TTinyErr", args{ct.Add(-time.Second * 59), time.Minute}, true},
		{"TFutureCreated", args{ct.Add(-time.Minute), time.Minute}, false},
		{"TInTime", args{ct.Add(time.Hour), time.Minute}, true},
		{"TNeedUpdate", args{ct.Add(time.Hour * 19), time.Minute}, true},
		{"TNearlyExpire", args{ct.Add(time.Hour * 24).Add(-time.Second), time.Minute}, true},
		{"TExpired", args{ct.Add(time.Hour * 25), time.Minute}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &AgentCertificateV1{
				CreateTime: ct,
				ExpireTime: et,
			}

			got := c.IsValidInTime(tt.args.t, tt.args.errTolerance)
			if got != tt.want {
				t.Errorf("IsValidInTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentCertificateV1_IsNeedUpdate(t *testing.T) {
	ct := time.Now()
	et := ct.Add(time.Hour * 24)

	type args struct {
		t       time.Time
		bufTerm time.Duration
	}
	tests := []struct {
		name string

		args args
		want bool
	}{
		{"TFresh", args{ct.Add(time.Second * 59), time.Hour*6}, false},
		{"TYoung", args{et.Add(-time.Hour*6), time.Hour*6}, false},
		{"TOld", args{et.Add(-time.Hour), time.Hour*6}, true},
		{"TExpired", args{et.Add(time.Minute), time.Hour*6}, true},
	}
	for _, tt := range tests {
		c := &AgentCertificateV1{
			CreateTime: ct,
			ExpireTime: et,
		}

		if got := c.IsNeedUpdate(tt.args.t, tt.args.bufTerm); got != tt.want {
			t.Errorf("IsNeedUpdate() = %v, want %v", got, tt.want)
		}
	}
}

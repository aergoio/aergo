/**
 *  @file
 *  @copyright defined in aergo/LICENSE.txt
 */

package component

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/aergoio/aergo-actor/actor"
	"github.com/aergoio/aergo-actor/mailbox"
	metrics "github.com/rcrowley/go-metrics"
)

type mailboxLogger struct {
	status metrics.Registry
}

func newMailBoxLogger(m metrics.Registry) mailbox.Producer {
	return mailbox.Unbounded(&mailboxLogger{status: m})
}

func parseMessage(msg interface{}) (string, string) {
	typeName := reflect.TypeOf(msg).String()
	envelop, ok := msg.(*actor.MessageEnvelope)
	if ok {
		qtime := envelop.GetHeader("qtime")
		if qtime == "" {
			envelop.SetHeader("qtime", fmt.Sprintf("%d", time.Now().UnixNano()))
		}
		//typeName = typeName + "/" + reflect.TypeOf(envelop.Message).String()
		typeName = reflect.TypeOf(envelop.Message).String()
		return typeName, qtime
	}
	return typeName, ""
}

func (m *mailboxLogger) MailboxStarted() {
}

func (m *mailboxLogger) MessagePosted(msg interface{}) {
	parseMessage(msg)
}

func (m *mailboxLogger) MessageReceived(msg interface{}) {
	typeName, qtime := parseMessage(msg)
	//c := metrics.GetOrRegisterCounter("recv/"+typeName, m.status)
	//c.Inc(1)
	if qtime != "" {
		posttime, _ := strconv.ParseInt(qtime, 10, 64)
		t := metrics.GetOrRegisterTimer("recv/"+typeName, m.status)
		t.Update(time.Duration(time.Now().UnixNano() - posttime))
	}
}

func (m *mailboxLogger) MailboxEmpty() {
}

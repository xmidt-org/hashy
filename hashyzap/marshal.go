package hashyzap

import (
	"strconv"

	"github.com/miekg/dns"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type MsgHdrMarshaler dns.MsgHdr

func (mhm MsgHdrMarshaler) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint16("id", mhm.Id)
	if opcode, exists := dns.OpcodeToString[mhm.Opcode]; exists {
		enc.AddString("opcode", opcode)
	} else {
		enc.AddString("opcode", "opcode"+strconv.Itoa(mhm.Opcode))
	}

	return nil
}

type QuestionMarshaler dns.Question

func (qm QuestionMarshaler) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", qm.Name)
	enc.AddString("class", dns.Class(qm.Qclass).String())
	enc.AddString("type", dns.Type(qm.Qtype).String())
	return nil
}

type QuestionsMarshaler []dns.Question

func (qm QuestionsMarshaler) MarshalLogArray(enc zapcore.ArrayEncoder) (err error) {
	for i := 0; err == nil && i < len(qm); i++ {
		err = enc.AppendObject(QuestionMarshaler(qm[i]))
	}

	return
}

type MsgMarshaler dns.Msg

func (mm *MsgMarshaler) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	err := MsgHdrMarshaler(mm.MsgHdr).MarshalLogObject(enc)
	if err == nil {
		err = enc.AddArray("questions", QuestionsMarshaler(mm.Question))
	}

	return err
}

func MsgField(key string, val *dns.Msg) zap.Field {
	return zap.Object(key, (*MsgMarshaler)(val))
}

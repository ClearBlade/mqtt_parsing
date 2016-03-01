package mqtt_parsing

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"
)

func asdf() { log.Println("ugh") }

func Test_LargePacketTCP(t *testing.T) {
	pay := make([]byte, 0xf0f00)
	for i, _ := range pay {
		pay[i] = 0x0f
	}
	tp, _ := NewTopicPath("one/two")
	pub := &Publish{
		Header: &StaticHeader{
			QOS:    0,
			Retain: false,
			DUP:    false,
		},
		Payload:   pay,
		Topic:     tp,
		MessageId: 69,
	}
	listener, err := net.Listen("tcp", ":63002")
	if err != nil {
		t.Error(err.Error())
	}

	wg := new(sync.WaitGroup)
	go func() {
		for i := 0; i < 100; i++ {
			go func() {
				wg.Add(1)
				con, err := net.Dial("tcp", ":63002")
				if err != nil {
					t.Error(err.Error())
				}
				encp := pub.Encode()

				con.Write(encp)

				<-time.After(time.Millisecond * 10)
				defer con.Close()

				wg.Done()
			}()
		}
	}()
	for i := 0; i < 100; i++ {
		con, err := listener.Accept()
		if err != nil {
			t.Error(err.Error())
			continue
		}
		_, err = DecodePacket(con)
		if err != nil {
			fmt.Println(err.Error)
			t.Error("got error " + err.Error())
		}
	}

	wg.Wait()
}

func Test_LargePacket(t *testing.T) {
	pay := make([]byte, 0xf0f00)
	for i, _ := range pay {
		pay[i] = 0x0f
	}
	tp, _ := NewTopicPath("One/Two")
	pub := &Publish{
		Header: &StaticHeader{
			QOS:    0,
			Retain: false,
			DUP:    false,
		},
		Payload:   pay,
		Topic:     tp,
		MessageId: 69,
	}
	wg := new(sync.WaitGroup)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			slc := bytes.NewBuffer(pub.Encode())
			_, err := DecodePacket(slc)
			if err != nil {

				fmt.Println(err)
				t.Error("wtf")
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkPacketDecode(b *testing.B) {
	tp, _ := NewTopicPath("Benchmark/Packet/Decode")
	//test connect decoding/encoding
	con := &Connect{
		ProtoName:      "MQTsdp",
		Version:        3,
		UsernameFlag:   false,
		PasswordFlag:   false,
		WillRetainFlag: false,
		WillQOS:        0,
		WillFlag:       false,
		CleanSeshFlag:  true,
		KeepAlive:      30,
		ClientId:       "13241",
		WillTopic:      tp,
		WillMessage:    "tommy this and tommy that and tommy ow's yer soul",
		Username:       "Username",
		Password:       "Password",
	}
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		slc := bytes.NewBuffer(con.Encode())
		pkt, err := DecodePacket(slc)

		if err != nil {
			fmt.Printf("pkt %+v, con %+v\n", pkt, con)
			b.Error("wtf")
		}
	}
}

func encodeTestHelper(toEncode Message) bool {
	byt := toEncode.Encode()
	buf := bytes.NewBuffer(byt)
	msg, err := DecodePacket(buf)
	if err != nil {
		log.Printf("error in here %+v\n", err.Error())
		return false
	}
	match := false
	switch msg.(type) {
	case *Connect:
		match = msg.Type() == CONNECT
	case *Connack:
		match = msg.Type() == CONNACK
	case *Publish:
		match = msg.Type() == PUBLISH
	case *Pubrec:
		match = msg.Type() == PUBREC
	case *Puback:
		match = msg.Type() == PUBACK
	case *Pubrel:
		match = msg.Type() == PUBREL
	case *Pubcomp:
		match = msg.Type() == PUBCOMP
	case *Subscribe:
		match = msg.Type() == SUBSCRIBE
	case *Suback:
		match = msg.Type() == SUBACK
	case *Unsubscribe:
		match = msg.Type() == UNSUBSCRIBE
	case *Unsuback:
		match = msg.Type() == UNSUBACK
	case *Pingreq:
		match = msg.Type() == PINGREQ
	case *Pingresp:
		match = msg.Type() == PINGRESP
	case *Disconnect:
		match = msg.Type() == DISCONNECT
	}
	if match != true {
		return false
	}
	return reflect.DeepEqual(toEncode, msg)
}

func Test_Connect(t *testing.T) {
	tp, _ := NewTopicPath("Clearblade/test/Connect")
	testPkt := &Connect{
		ProtoName:      "MQTsdp",
		Version:        3,
		UsernameFlag:   true,
		PasswordFlag:   true,
		WillRetainFlag: true,
		WillQOS:        0,
		WillFlag:       true,
		CleanSeshFlag:  true,
		KeepAlive:      30,
		ClientId:       "420",
		WillTopic:      tp,
		WillMessage:    "tommy this and tommy that and tommy ow's yer soul",
		Username:       "Username",
		Password:       "Password is my password",
	}
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode connect failed")
	}
}

func Test_Connack(t *testing.T) {
	testPkt := &Connack{

		ReturnCode: 0x04,
	}
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode connack failed")
	}
}

func Test_Publish(t *testing.T) {
	testPkt := &Publish{
		Header: &StaticHeader{
			QOS:    1,
			Retain: false,
			DUP:    false,
		},
		Payload:   []byte("tommy this and tommy that"),
		Topic:     TopicPath{Split: []string{"Crab", "Battle"}, Whole: "Crab/Battle"},
		MessageId: 69,
	}
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode publish failed")
	}
}

func Test_Publish2(t *testing.T) {
	testPkt := &Publish{
		Header: &StaticHeader{
			QOS:    2,
			Retain: false,
			DUP:    false,
		},
		Payload:   []byte("A thin red line of 'eroes"),
		Topic:     TopicPath{Split: []string{"First", "Second"}, Whole: "First/Second"},
		MessageId: 69,
	}
	dinger := testPkt.Encode()
	buf := bytes.NewBuffer(dinger)
	msg, err := DecodePacket(buf)
	if err != nil {
		t.Error(err.Error())
	}
	if msg.(*Publish).Header.QOS != testPkt.Header.QOS {
		t.Error("Encode/decode failed on test publish 2")
	}
}

func Test_Publish_WithUnicodeDecoding(t *testing.T) {
	pay := []byte("hello earth 😁, good evening")
	testPkt := &Publish{
		Header: &StaticHeader{
			QOS:    2,
			Retain: false,
			DUP:    false,
		},
		Payload:   pay,
		Topic:     TopicPath{Split: []string{"Crab", "Battle"}, Whole: "Crab/Battle"},
		MessageId: 69,
	}
	dinger := testPkt.Encode()
	buf := bytes.NewBuffer(dinger)
	msg, err := DecodePacket(buf)
	if err != nil {
		t.Error(err.Error())
	}
	if msg.(*Publish).Header.QOS != testPkt.Header.QOS {
		t.Error("Encode/decode failed on test publish 2")
	}
	if !bytes.Equal(msg.(*Publish).Payload, pay) {
		log.Println(pay)
		log.Println(msg.(*Publish).Payload)

		log.Println(string(pay))
		log.Println(string(msg.(*Publish).Payload))

		t.Error("Invalid encoding")
	}
}

func Test_Puback(t *testing.T) {
	testPkt := &Puback{
		MessageId: 0xbeef,
	}
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode puback failed")
	}
}

func Test_Pubrec(t *testing.T) {
	testPkt := &Pubrec{
		MessageId: 0xbeef,
	}
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode pubrec failed")
	}
}

func Test_Pubrel(t *testing.T) {
	testPkt := &Pubrel{
		MessageId: 0xbeef,
		Header: &StaticHeader{
			QOS:    1,
			Retain: false,
			DUP:    false,
		},
	}
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode pubrel failed")
	}
}

func Test_Pubcomp(t *testing.T) {
	testPkt := &Pubcomp{
		MessageId: 0xbeef,
	}
	if !encodeTestHelper(testPkt) {

		t.Error("encode/decode pubcomp failed")
	}
}

func Test_Subscribe(t *testing.T) {
	testPkt := &Subscribe{
		MessageId: 0xbeef,
		Header: &StaticHeader{
			QOS:    1,
			Retain: false,
			DUP:    false,
		},
		Subscriptions: []TopicQOSTuple{
			TopicQOSTuple{
				Qos:   0,
				Topic: TopicPath{Split: []string{"one", "two"}, Whole: "one/two"},
			},
		},
	}
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode subscribe failed")
	}
}

func Test_Suback(t *testing.T) {
	testPkt := &Suback{
		MessageId: 0xbeef,
		Qos:       []uint8{0, 0, 1},
	}
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode suback failed")
	}
}

func Test_UnSubscribe(t *testing.T) {
	testPkt := &Unsubscribe{
		MessageId: 0xbeef,
		Header: &StaticHeader{
			QOS:    1,
			Retain: false,
			DUP:    false,
		},
		Topics: []TopicQOSTuple{
			TopicQOSTuple{
				Qos:   0,
				Topic: TopicPath{Split: []string{"One", "Two"}, Whole: "One/Two"},
			},
		},
	}
	if !encodeTestHelper(testPkt) {
		t.Error("encode/decode unsubscribe failed")
	}
}

func Test_Unsuback(t *testing.T) {
	testPkt := &Unsuback{
		MessageId: 0xbeef,
	}
	if !encodeTestHelper(testPkt) {

		t.Error("encode/decode unsuback failed")
	}
}

func Test_PingReq(t *testing.T) {
	testPkt := &Pingreq{}
	if !encodeTestHelper(testPkt) {

		t.Error("encode/decode pingreq failed")
	}
}

func Test_PingResp(t *testing.T) {
	testPkt := &Pingresp{}
	if !encodeTestHelper(testPkt) {

		t.Error("encode/decode pingresp failed")
	}
}

func Test_Disconnect(t *testing.T) {
	testPkt := &Disconnect{}
	if !encodeTestHelper(testPkt) {

		t.Error("encode/decode disconnect failed")
	}
}

func Test_encodeLength(t *testing.T) {
	test := func(testval, expecField uint32, expecLeng uint8, t *testing.T) {
		fmtStr := "invalid response from encodeLength field %b leng %d, expected field %b expected value %d\n"
		leng, field := encodeLength(testval)
		if field != expecField || leng != expecLeng {
			t.Error(fmt.Sprintf(fmtStr, field, leng, expecField, expecLeng))
		}
	}

	test(0, 0x0, 1, t)
	test(1, 0x1, 1, t)
	test(127, 0x7f, 1, t)
	test(128, 0x8001, 2, t)
	test(16383, 0xff7f, 2, t)
	test(16384, 0x808001, 3, t)
	test(2097151, 0xffff7f, 3, t)
	test(2097152, 0x80808001, 4, t)
	test(268435455, 0xffffff7f, 4, t)
}

func Test_DecodeLength(t *testing.T) {
	tst := func(tst uint32, t *testing.T) {
		_, enclen := encodeLength(tst)
		var byteee [4]byte
		byteee[0] = byte(enclen >> 24)
		byteee[1] = byte(enclen >> 16)
		byteee[2] = byte(enclen >> 8)
		byteee[3] = byte(enclen)
		if res := decodeLen(byteee[:]); res != tst {
			t.Error(fmt.Sprintf("expected %d and got %d\n", tst, res))
		}
	}

	test := func(expecField, testval uint32, expecLeng uint8, t *testing.T) {
		fmtStr := "invalid response from encodeLength field %b leng %d, expected field %b expected value %d\n"
		var blah [4]byte
		blah[0] = byte(testval >> 24)
		blah[1] = byte(testval >> 16)
		blah[2] = byte(testval >> 8)
		blah[3] = byte(testval)

		field := decodeLen(blah[:])
		if field != expecField {
			t.Error(fmt.Sprintf(fmtStr, field, 0, expecField, expecLeng))
		}
	}
	tst(986889, t)
	tst(0, t)
	tst(1, t)
	tst(127, t)
	tst(128, t)
	tst(16383, t)
	tst(16384, t)
	tst(209715, t)
	tst(2097152, t)
	tst(268435455, t)
	test(0, 0x0, 1, t)
	test(1, 0x1, 1, t)
	test(127, 0x7f, 1, t)
	test(128, 0x8001, 2, t)
	test(16383, 0xff7f, 2, t)
	test(16384, 0x808001, 3, t)
	test(2097151, 0xffff7f, 3, t)
	test(2097152, 0x80808001, 4, t)
	test(268435455, 0xffffff7f, 4, t)

}

func decodeLen(field []byte) uint32 {
	//sadly I have to ape decoding length
	multiplier := uint32(1)
	length := uint32(0) //signed for great justice?
	digit := byte(0x80)
	rdr := bytes.NewBuffer(field)

	//since we're writing the byte pattern to a 4 byte slice, no matter what the actual size, we have to skip the leftmost empty bytes
	var b [1]byte
	steps := 1
	for (digit & 0x80) != 0 {
		_, err := io.ReadFull(rdr, b[:])
		if err != nil {
			log.Println(digit, steps)
			panic(err.Error())
		}
		if b[0] == 0 {
			if steps == 4 {
				return 0
			} else {
				steps++
				continue
			}
		}
		steps++
		digit = b[0]

		length += uint32(digit&0x7f) * multiplier
		multiplier *= 128

	}
	return length
}

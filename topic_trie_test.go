package gmqtt

import (
	"github.com/DrmagicE/gmqtt/pkg/packets"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testTopicMatch = []struct {
	subTopic string //subscribe topic
	topic    string //publish topic
	isMatch  bool
}{
	{subTopic: "#", topic: "/abc/def", isMatch: true},
	{subTopic: "/a", topic: "a", isMatch: false},
	{subTopic: "a/#", topic: "a", isMatch: true},
	{subTopic: "+", topic: "/a", isMatch: false},

	{subTopic: "a/", topic: "a", isMatch: false},
	{subTopic: "a/+", topic: "a/123/4", isMatch: false},
	{subTopic: "a/#", topic: "a/123/4", isMatch: true},

	{subTopic: "/a/+/+/abcd", topic: "/a/dfdf/3434/abcd", isMatch: true},
	{subTopic: "/a/+/+/abcd", topic: "/a/dfdf/3434/abcdd", isMatch: false},
	{subTopic: "/a/+/abc/", topic: "/a/dfdf/abc/", isMatch: true},
	{subTopic: "/a/+/abc/", topic: "/a/dfdf/abc", isMatch: false},
	{subTopic: "/a/+/+/", topic: "/a/dfdf/", isMatch: false},
	{subTopic: "/a/+/+", topic: "/a/dfdf/", isMatch: true},
	{subTopic: "/a/+/+/#", topic: "/a/dfdf/", isMatch: true},
}

var topicMatchQosTest = []struct {
	topics     []packets.Topic
	matchTopic struct {
		name string // matched topic name
		qos  uint8  // matched qos
	}
}{
	{
		topics: []packets.Topic{
			{
				Name: "a/b",
				Qos:  packets.QOS_1,
			},
			{
				Name: "a/#",
				Qos:  packets.QOS_2,
			},
			{
				Name: "a/+",
				Qos:  packets.QOS_0,
			},
		},
		matchTopic: struct {
			name string
			qos  uint8
		}{
			name: "a/b",
			qos:  packets.QOS_2,
		},
	},
}

var testSubscribeAndFind = struct {
	subTopics  map[string][]packets.Topic // subscription
	findTopics map[string][]struct {      //key by clientID
		exist     bool
		topicName string
		wantQos   uint8
	}
}{
	subTopics: map[string][]packets.Topic{
		"cid1": {
			{packets.QOS_1, "t1/t2/+"},
			{packets.QOS_2, "t1/t2/"},
			{packets.QOS_0, "t1/t2/cid1"},
		},
		"cid2": {
			{packets.QOS_2, "t1/t2/+"},
			{packets.QOS_1, "t1/t2/"},
			{packets.QOS_0, "t1/t2/cid2"},
		},
	},
	findTopics: map[string][]struct { //key by clientID
		exist     bool
		topicName string
		wantQos   uint8
	}{
		"cid1": {
			{exist: true, topicName: "t1/t2/+", wantQos: packets.QOS_1},
			{exist: true, topicName: "t1/t2/", wantQos: packets.QOS_2},
			{exist: false, topicName: "t1/t2/cid2"},
			{exist: false, topicName: "t1/t2/cid3"},
		},
		"cid2": {
			{exist: true, topicName: "t1/t2/+", wantQos: packets.QOS_2},
			{exist: true, topicName: "t1/t2/", wantQos: packets.QOS_1},
			{exist: false, topicName: "t1/t2/cid1"},
		},
	},
}

var testUnsubscribe = struct {
	subTopics   map[string][]packets.Topic //key by clientID
	unsubscribe map[string][]string          // clientID => topic name
	afterUnsub  map[string][]struct {      // test after unsubscribe, key by clientID
		exist     bool
		topicName string
		wantQos   uint8
	}
}{
	subTopics: map[string][]packets.Topic{
		"cid1": {
			{packets.QOS_1, "t1/t2/t3"},
			{packets.QOS_2, "t1/t2"},
		},
		"cid2": {
			{packets.QOS_2, "t1/t2/t3"},
			{packets.QOS_1, "t1/t2"},
		},
	},
	unsubscribe: map[string][]string{
		"cid1": {"t1/t2/t3","t4/t5"},
		"cid2": {"t1/t2/t3"},
	},
	afterUnsub: map[string][]struct { // test after unsubscribe
		exist     bool
		topicName string
		wantQos   uint8
	}{
		"cid1": {
			{exist: false, topicName: "t1/t2/t3"},
			{exist: true, topicName: "t1/t2", wantQos: packets.QOS_2},
		},
		"cid2": {
			{exist: false, topicName: "t1/t2/+"},
			{exist: true, topicName: "t1/t2", wantQos: packets.QOS_1},
		},
	},
}

func TestTopicTrie_matchedClients(t *testing.T) {
	a := assert.New(t)
	for _, v := range testTopicMatch {
		trie := newTopicTrie()
		topic := packets.Topic{Qos: 1, Name: v.subTopic}
		trie.subscribe("cid", topic)
		qos := trie.getMatchedClients(v.topic)
		if v.isMatch {
			a.Equal(qos["cid"], topic.Qos)
		} else {
			_, ok := qos["cid"]
			a.False(ok)
		}
	}
}

func TestTopicTrie_matchedClients_Qos(t *testing.T) {
	a := assert.New(t)
	for _, v := range topicMatchQosTest {
		trie := newTopicTrie()
		for _, tt := range v.topics {
			trie.subscribe("cid", tt)
		}
		rs := trie.getMatchedClients(v.matchTopic.name)
		a.Equal(v.matchTopic.qos, rs["cid"])
	}
}
func TestTopicTrie_subscribeAndFind(t *testing.T) {
	a := assert.New(t)
	trie := newTopicTrie()
	for cid, v := range testSubscribeAndFind.subTopics {
		for _, topic := range v {
			trie.subscribe(cid, topic)
		}
	}
	for cid, v := range testSubscribeAndFind.findTopics {
		for _, tt := range v {
			node := trie.find(tt.topicName)
			if tt.exist {
				a.Equal(tt.wantQos, node.clients[cid])
			} else {
				if node != nil {
					_, ok := node.clients[cid]
					a.False(ok)
				}
			}
		}
	}
}

func TestTopicTrie_unsubscribe(t *testing.T) {
	a := assert.New(t)
	trie := newTopicTrie()
	for cid, v := range testUnsubscribe.subTopics {
		for _, topic := range v {
			trie.subscribe(cid, topic)
		}
	}
	for cid, v := range testUnsubscribe.unsubscribe {
		for _, tt := range v {
			trie.unsubscribe(cid, tt)
		}
	}
	for cid,v := range testUnsubscribe.afterUnsub {
		for _, tt := range v {
			matched := trie.getMatchedClients(tt.topicName)
			if tt.exist {
				a.Equal(matched[cid],tt.wantQos)
			} else {
				a.Equal(0,len(matched))
			}
		}
	}
}





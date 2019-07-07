package gmqtt

import (
	"github.com/DrmagicE/gmqtt/pkg/packets"
	"strings"
)

// topicTrie 主题树
type topicTrie = topicNode
type children = map[string]*topicNode

// topicNode 表示主题树中的一个节点
type topicNode struct {
	children children
	clients  map[string]uint8 // clientID => qos
	parent *topicNode  // pointer of parent node
	topicName string
}

// newTopicTrie 返回一个指向初始化后的TrieTree指针
func newTopicTrie() *topicTrie {
	return newNode()
}

// newNode 创建一个节点
func newNode() *topicNode {
	return &topicNode{
		children: children{},
		clients:  make(map[string]uint8),
	}
}
func (t *topicNode) newChild() *topicNode {
	return &topicNode{
		children: children{},
		clients:  make(map[string]uint8),
		parent:t,
	}
}

// subscribe 订阅topic,返回是否新的订阅
func (t *topicTrie) subscribe(clientID string, topic packets.Topic) (bool, *topicNode) {
	topicSlice := strings.Split(topic.Name, "/")
	var pNode = t
	var isNew bool
	for _, lv := range topicSlice {
		if _, ok := pNode.children[lv]; !ok {
			pNode.children[lv] = t.newChild()
		}
		pNode = pNode.children[lv]
	}
	if _, ok := pNode.clients[clientID]; !ok {
		isNew = true
	}
	pNode.clients[clientID] = topic.Qos
	pNode.topicName = topic.Name
	return isNew, pNode
}

// find找到某节点
func (t *topicTrie) find(topicName string) *topicNode {
	topicSlice := strings.Split(topicName, "/")
	var pNode = t
	for _, lv := range topicSlice {
		if _, ok := pNode.children[lv]; ok {
			pNode = pNode.children[lv]
		} else {
			return nil
		}
	}
	return pNode
}

// unsubscribe 取消订阅
func (t *topicTrie) unsubscribe(clientID string, topicName string) {
	topicSlice := strings.Split(topicName, "/")
	l := len(topicSlice)
	var pNode = t
	for _, lv := range topicSlice[0 : l-1] {
		if _, ok := pNode.children[lv]; ok {
			pNode = pNode.children[lv]
		} else {
			return
		}
	}
	if n, ok := pNode.children[topicSlice[l-1]]; ok {
		delete(n.clients, clientID)
		if len(n.clients) == 0 && len(n.children) == 0 { // empty leaf
			delete(pNode.children, topicSlice[l-1])
		}
	}
}

// setQos
func setQos(node *topicNode, clients map[string]uint8) {
	for cid, qos := range node.clients {
		if qos >= clients[cid] {
			clients[cid] = qos
		}
	}
}

// matchTopic
func (t *topicTrie) matchTopic(topicSlice []string, qos map[string]uint8) {
	endFlag := len(topicSlice) == 1
	if cnode := t.children["#"]; cnode != nil {
		setQos(cnode, qos)
	}
	if cnode := t.children["+"]; cnode != nil {
		if endFlag {
			setQos(cnode, qos)
			if n := cnode.children["#"]; n != nil {
				setQos(n, qos)
			}
		} else {
			cnode.matchTopic(topicSlice[1:], qos)
		}
	}
	if cnode := t.children[topicSlice[0]]; cnode != nil {
		if endFlag {
			setQos(cnode, qos)
			if n := cnode.children["#"]; n != nil {
				setQos(n, qos)
			}
		} else {
			cnode.matchTopic(topicSlice[1:], qos)
		}
	}
}

// getMatchedClients 返回符合的客户端ID，匹配topic的QOS要最高
func (t *topicTrie) getMatchedClients(topicName string) map[string]uint8 {
	topicLv := strings.Split(topicName, "/")
	qos := make(map[string]uint8)
	t.matchTopic(topicLv, qos)
	return qos
}

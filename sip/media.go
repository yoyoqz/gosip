package sipapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/panjjo/gosip/m"
)

type Strategy interface {
	GetNext() *m.MediaServer
	HealthCheck()
}

var Instance Strategy

func GetStrategy(name string) Strategy {
	if Instance == nil {
		if name == "1" {
			Instance = NewRoundRobin()
		} else {
			Instance = NewWeightRoundRobinBalance()
		}
	}

	return Instance
}

type RoundRobin struct {
	index int

	nodes []*Node
}

type Node struct {
	server  *m.MediaServer
	IsAlive bool
}

func NewRoundRobin() *RoundRobin {
	w := &RoundRobin{}

	for _, v := range config.Media {
		n := &Node{}
		n.server = &v
		n.IsAlive = true

		w.nodes = append(w.nodes, n)
	}

	return w
}

func (r *RoundRobin) GetNext() *m.MediaServer {
	now := (r.index) % len(config.Media)
	r.index++

	return &config.Media[now]
}

func (r *RoundRobin) HealthCheck() {
	for i := range r.nodes {
		//r.rn[i].server.HTTP
		if CheckSiteStatus(r.nodes[i].server.HTTP) {
			r.nodes[i].IsAlive = true
		} else {
			r.nodes[i].IsAlive = false
		}
	}
}

type WeightNode struct {
	Node
	weight          int // 权重
	currentWeight   int //节点当前权重
	effectiveWeight int //有效权重
}

type WeightRoundRobinBalance struct {
	nodes []*WeightNode
}

func NewWeightRoundRobinBalance() *WeightRoundRobinBalance {
	w := &WeightRoundRobinBalance{}

	for _, v := range config.Media {
		n := &WeightNode{}
		n.IsAlive = true
		n.server = &v
		n.weight = v.Weight

		w.nodes = append(w.nodes, n)
	}

	return w
}

func (wb *WeightRoundRobinBalance) HealthCheck() {
	for i := range wb.nodes {
		if CheckSiteStatus(wb.nodes[i].server.HTTP) {
			wb.nodes[i].IsAlive = true
		} else {
			wb.nodes[i].IsAlive = false
		}
	}
}

func (wb *WeightRoundRobinBalance) GetNext() *m.MediaServer {
	total := 0
	var best *WeightNode //该次最优的ip
	for i := 0; i < len(wb.nodes); i++ {
		w := wb.nodes[i]
		if !w.IsAlive {
			continue
		}

		//统计所有有效权重之和
		total += w.effectiveWeight
		//变更节点临时权重为的节点临时权重+节点有效权重
		w.currentWeight += w.effectiveWeight
		//有效权重默认与权重相同，通讯异常时-1, 通讯成功+1，直到恢复到weight大小
		if w.effectiveWeight < w.weight {
			w.effectiveWeight++
		}
		//选择最大临时权重点节点
		if best == nil || w.currentWeight > best.currentWeight {
			best = w
		}
	}
	if best == nil {
		return nil
	}
	//变更临时权重为 临时权重-有效权重之和
	best.currentWeight -= total
	return best.server
}

func CheckSiteStatus(url string) bool {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Site not OK. Response code: %d\n", resp.StatusCode)
		return false
	}
	return true
}

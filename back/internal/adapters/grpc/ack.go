package grpc

import (
	"github.com/brice-74/sensorflow/internal/adapters/grpc/proto"
)

type AckAggregator struct {
	Accepted      uint64
	Rejected      uint64
	RejectDetails RejectedAggregator
}

func NewAckAggregator() *AckAggregator {
	return &AckAggregator{
		RejectDetails: make(RejectedAggregator),
	}
}

func (agg *AckAggregator) AddAccepted(n uint64) {
	agg.Accepted += n
}

func (agg *AckAggregator) AddRejected(sensorID string, code proto.RejectionCode, reason ...string) {
	agg.Rejected++
	agg.RejectDetails.Add(sensorID, code, reason...)
}

func (agg *AckAggregator) ToProto(mode proto.AckMode) *proto.IngestAck {
	ack := &proto.IngestAck{
		Accepted: agg.Accepted,
		Rejected: agg.Rejected,
	}

	if mode == proto.AckMode_ACK_MODE_FULL {
		ack.RejectedDetails = agg.RejectDetails.ToProto()
	}

	return ack
}

type RejectedAggregator map[string]map[proto.RejectionCode]*RejectedCount

type RejectedCount struct {
	Count  uint64
	Reason string
}

func (ra RejectedAggregator) Add(sensorID string, code proto.RejectionCode, reason ...string) {
	if ra[sensorID] == nil {
		ra[sensorID] = make(map[proto.RejectionCode]*RejectedCount)
	}

	entry, exists := ra[sensorID][code]
	if !exists {
		msg := code.String()
		if len(reason) > 0 && reason[0] != "" {
			msg = reason[0]
		}
		entry = &RejectedCount{
			Count:  0,
			Reason: msg,
		}
		ra[sensorID][code] = entry
	}

	entry.Count++
}

func (ra RejectedAggregator) ToProto() []*proto.RejectedDetail {
	details := make([]*proto.RejectedDetail, 0)
	for sensorID, codes := range ra {
		for code, entry := range codes {
			details = append(details, &proto.RejectedDetail{
				SensorId: sensorID,
				Code:     code,
				Reason:   entry.Reason,
				Count:    entry.Count,
			})
		}
	}
	return details
}

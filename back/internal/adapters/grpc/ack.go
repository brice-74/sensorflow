package grpc

import (
	"github.com/brice-74/sensorflow/internal/adapters/grpc/proto"
)

type key struct {
	sensorID string
	code     proto.RejectionCode
	reason   string
}

type RejectedAggregator map[key]uint64

func NewRejectedAggregator() RejectedAggregator {
	return make(RejectedAggregator)
}

func (ra RejectedAggregator) AddN(sensorID string, code proto.RejectionCode, reason string, n uint64) {
	k := key{sensorID, code, reason}
	ra[k] += n
}

func (ra RejectedAggregator) Add(sensorID string, code proto.RejectionCode, reason string) {
	ra.AddN(sensorID, code, reason, 1)
}

func (ra RejectedAggregator) ToProto() []*proto.RejectedDetail {
	result := make([]*proto.RejectedDetail, 0, len(ra))
	for k, count := range ra {
		result = append(result, &proto.RejectedDetail{
			SensorId: k.sensorID,
			Code:     k.code,
			Reason:   k.reason,
			Count:    count,
		})
	}
	return result
}

type AckAggregator struct {
	accepted        uint64
	rejected        uint64
	details         RejectedAggregator
	logicalStatus   proto.LogicalIngestStatus
	technicalStatus proto.TechnicalIngestStatus
	message         *string

	mode proto.AckMode
}

func NewAckAggregator(mode proto.AckMode) *AckAggregator {
	var details RejectedAggregator
	if mode == proto.AckMode_ACK_MODE_FULL {
		details = NewRejectedAggregator()
	}
	return &AckAggregator{
		details: details,
		mode:    mode,
	}
}
func (agg *AckAggregator) AddAccepted(n uint64) *AckAggregator {
	agg.accepted += n
	return agg
}

func (agg *AckAggregator) AddRejected(sensorID string, code proto.RejectionCode, reason string) *AckAggregator {
	return agg.AddRejectedN(sensorID, code, reason, 1)
}

func (agg *AckAggregator) AddRejectedN(sensorID string, code proto.RejectionCode, reason string, n uint64) *AckAggregator {
	agg.rejected += n
	if agg.details != nil {
		agg.details.AddN(sensorID, code, reason, n)
	}
	return agg
}

func (agg *AckAggregator) SetLogicalStatus(status proto.LogicalIngestStatus) *AckAggregator {
	agg.logicalStatus = status
	return agg
}

func (agg *AckAggregator) SetTechnicalStatus(status proto.TechnicalIngestStatus) *AckAggregator {
	agg.technicalStatus = status
	return agg
}

func (agg *AckAggregator) SetMessage(msg string) *AckAggregator {
	agg.message = &msg
	return agg
}

func (agg *AckAggregator) ToProto() *proto.IngestAck {
	if agg.mode == proto.AckMode_ACK_MODE_NONE {
		return nil
	}

	ack := &proto.IngestAck{
		Accepted:        agg.accepted,
		Rejected:        agg.rejected,
		LogicalStatus:   agg.logicalStatus,
		TechnicalStatus: agg.technicalStatus,
		Message:         agg.message,
	}

	if agg.logicalStatus == proto.LogicalIngestStatus_LOGICAL_INGEST_STATUS_UNSPECIFIED {
		agg.computeLogicalStatus()
	}

	if agg.mode == proto.AckMode_ACK_MODE_FULL && agg.details != nil {
		ack.RejectedDetails = agg.details.ToProto()
	}

	return ack
}

func (agg *AckAggregator) computeLogicalStatus() {
	switch {
	case agg.accepted > 0 && agg.rejected == 0:
		agg.logicalStatus = proto.LogicalIngestStatus_LOGICAL_INGEST_STATUS_OK
	case agg.accepted > 0 && agg.rejected > 0:
		agg.logicalStatus = proto.LogicalIngestStatus_LOGICAL_INGEST_STATUS_PARTIAL
	case agg.accepted == 0 && agg.rejected > 0:
		agg.logicalStatus = proto.LogicalIngestStatus_LOGICAL_INGEST_STATUS_REJECTED
	default:
		agg.logicalStatus = proto.LogicalIngestStatus_LOGICAL_INGEST_STATUS_UNSPECIFIED
	}
}

package webapp

import (
	"container/list"
	"time"
)

// SlidingWindowLimiter 滑动窗口限流器
type SlidingWindowLimiter struct {
	baseLimiter
	segments map[string][]*struct {
		count     int
		timestamp time.Time
	}
	options *SlidingWindowOptions
}

func NewSlidingWindowLimiter(options *SlidingWindowOptions) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		baseLimiter: baseLimiter{
			queue: make(map[string]*list.List),
		},
		segments: make(map[string][]*struct {
			count     int
			timestamp time.Time
		}),
		options: options,
	}
}

func (l *SlidingWindowLimiter) TryAcquire(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	segments := l.segments[key]

	// 清理过期的segments
	cutoff := now.Add(-l.options.Window)
	validSegments := make([]*struct {
		count     int
		timestamp time.Time
	}, 0)

	totalCount := 0
	for _, seg := range segments {
		if seg.timestamp.After(cutoff) {
			validSegments = append(validSegments, seg)
			totalCount += seg.count
		}
	}

	if totalCount >= l.options.PermitLimit {
		if l.options.QueueLimit > 0 {
			queue := l.getOrCreateQueue(key)
			if queue.Len() < l.options.QueueLimit {
				queue.PushBack(now)
				nextTime := validSegments[0].timestamp.Add(l.options.Window)
				return false, nextTime.Sub(now)
			}
		}
		nextTime := validSegments[0].timestamp.Add(l.options.Window)
		return false, nextTime.Sub(now)
	}

	// 添加新请求到当前segment
	segmentDuration := l.options.Window / time.Duration(l.options.SegmentsPerWindow)
	currentSegmentTime := now.Truncate(segmentDuration)

	var currentSegment *struct {
		count     int
		timestamp time.Time
	}

	for _, seg := range validSegments {
		if seg.timestamp.Equal(currentSegmentTime) {
			currentSegment = seg
			break
		}
	}

	if currentSegment == nil {
		currentSegment = &struct {
			count     int
			timestamp time.Time
		}{
			timestamp: currentSegmentTime,
		}
		validSegments = append(validSegments, currentSegment)
	}

	currentSegment.count++
	l.segments[key] = validSegments

	return true, 0
}

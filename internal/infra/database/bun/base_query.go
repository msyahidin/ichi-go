package bun

import (
	"time"

	"github.com/uptrace/bun"
)

type QueryScope func(*bun.SelectQuery) *bun.SelectQuery

func WhereStatus(status string) QueryScope {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("status = ?", status)
	}
}

func WhereActive() QueryScope {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("status = ?", "active")
	}
}

func WhereNotDeleted() QueryScope {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("deleted_at IS NULL")
	}
}

func WhereCreatedBetween(startDate, endDate time.Time) QueryScope {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("created_at BETWEEN ? AND ?", startDate, endDate)
	}
}

func WhereCreatedToday() QueryScope {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		today := time.Now().Truncate(24 * time.Hour)
		tomorrow := today.Add(24 * time.Hour)
		return q.Where("created_at >= ? AND created_at < ?", today, tomorrow)
	}
}

func WhereCreatedThisMonth() QueryScope {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		now := time.Now()
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, 0)
		return q.Where("created_at >= ? AND created_at < ?", startOfMonth, endOfMonth)
	}
}

func OrderByLatest() QueryScope {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Order("created_at DESC")
	}
}

func OrderByOldest() QueryScope {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Order("created_at ASC")
	}
}

func Limit(limit int) QueryScope {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Limit(limit)
	}
}

func Paginate(page, perPage int) QueryScope {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		offset := (page - 1) * perPage
		return q.Limit(perPage).Offset(offset)
	}
}

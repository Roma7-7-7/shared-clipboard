package dal

import "time"

type (
	User struct {
		ID           uint64
		Name         string
		Password     string
		PasswordSalt string
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}

	SessionFilter struct {
		userID          uint64
		limit           int
		name            string
		sortBy          string
		sortByDirection string
		offset          int
	}

	Session struct {
		ID        uint64
		Name      string
		UserID    uint64
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	Clipboard struct {
		SessionID   uint64
		ContentType string
		Content     []byte
		UpdatedAt   time.Time
	}
)

func (s SessionFilter) UserID() uint64 {
	return s.userID
}

func (s SessionFilter) Limit() int {
	return s.limit
}

func (s SessionFilter) Name() string {
	return s.name
}

func (s SessionFilter) SortBy() string {
	return s.sortBy
}

func (s SessionFilter) SortByDirection() string {
	return s.sortByDirection
}

func (s SessionFilter) Offset() int {
	return s.offset
}

func WithName(name string) func(SessionFilter) SessionFilter {
	return func(f SessionFilter) SessionFilter {
		f.name = name
		return f
	}
}

func WithSortByName(desc bool) func(SessionFilter) SessionFilter {
	return func(f SessionFilter) SessionFilter {
		f.sortBy = "name"
		if desc {
			f.sortByDirection = "DESC"
		} else {
			f.sortByDirection = "ASC"
		}
		return f
	}
}

func WithSortByUpdateAt(desc bool) func(SessionFilter) SessionFilter {
	return func(f SessionFilter) SessionFilter {
		f.sortBy = "updated_at"
		if desc {
			f.sortByDirection = "DESC"
		} else {
			f.sortByDirection = "ASC"
		}
		return f
	}
}

func WithOffset(offset int) func(SessionFilter) SessionFilter {
	return func(f SessionFilter) SessionFilter {
		f.offset = offset
		return f
	}
}

func NewSessionFilter(userID uint64, limit int, opts ...func(SessionFilter) SessionFilter) SessionFilter {
	f := SessionFilter{
		userID:          userID,
		limit:           limit,
		name:            "",
		sortBy:          "updated_at",
		sortByDirection: "ASC",
		offset:          0,
	}

	for _, opt := range opts {
		f = opt(f)
	}

	return f
}

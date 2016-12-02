	"github.com/m3db/m3db/clock"
	"github.com/m3db/m3x/instrument"
	"github.com/m3db/m3db/ratelimit"
// Options represents the options for filesystem persistence
	// SetClockOptions sets the clock options
	SetClockOptions(value clock.Options) Options

	// ClockOptions returns the clock options
	ClockOptions() clock.Options

	// SetRateLimitOptions sets the rate limit options
	SetRateLimitOptions(value ratelimit.Options) Options

	// RateLimitOptions returns the rate limit options
	RateLimitOptions() ratelimit.Options

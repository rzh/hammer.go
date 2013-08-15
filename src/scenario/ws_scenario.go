package scenario

import (
	"errors"
	"log"
	"math/rand"
	"strconv"
	// "sync/atomic"
	"time"
)

type WorldServerScenario struct {
	SessionScenario
	SessionAmount int

	_hex_cache [][]chan int64
}

func (ss *WorldServerScenario) InitFromCode() {
	ss._sessions = make([]*Session, ss.SessionAmount)

	ss._hex_cache = make([][]chan int64, ss.SessionAmount)
	for i := 0; i < ss.SessionAmount; i++ {
		ss._hex_cache[i] = make([]chan int64, ss.SessionAmount)
		for j := 0; j < ss.SessionAmount; j++ {
			ss._hex_cache[i][j] = make(chan int64, 1)
			ss._hex_cache[i][j] <- 0
		}
	}
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < ss.SessionAmount; i++ {
		// k := i
		x := rand.Intn(ss.SessionAmount)
		y := rand.Intn(ss.SessionAmount)
		ss.addSession([]GenSession{
			GenSession(func() (float32, GenCall, GenCallBack) {
				var (
					v int64
				)

				stamp := strconv.Itoa(rand.Intn(80000))
				return 0,
					GenCall(func(ps ...string) (_m, _t, _u, _b string) {
						// log.Println("1 call gen", k, v)

						v = <-ss._hex_cache[x][y]
						return "POST", "REST",
							"http://localhost:58888/v1/transaction",
							"{\"set\":{\"obj\":[{\"x\":" + strconv.Itoa(x) +
								",\"y\":" + strconv.Itoa(y) + ",\"version\":" +
								strconv.FormatInt(v, 10) + ",\"data\":{\"dude\":" +
								stamp + "}}]}}"
					}),
					GenCallBack(func(se *Session, st int, storage []byte) {
						se.StepLock <- se.State + st
						ss._hex_cache[x][y] <- v + 1
					})
			}),
			GenSession(func() (float32, GenCall, GenCallBack) {
				var (
					v int64
				)
				stamp := strconv.Itoa(rand.Intn(80000))
				return 50,
					GenCall(func(ps ...string) (_m, _t, _u, _b string) {
						// log.Println("2 call gen", x, v, stamp)
						v = <-ss._hex_cache[x][y]
						return "POST", "REST",
							"http://localhost:58888/v1/transaction",
							"{\"set\":{\"obj\":[{\"x\":" + strconv.Itoa(x) +
								",\"y\":" + strconv.Itoa(y) + ",\"version\":" +
								strconv.FormatInt(v, 10) + ",\"data\":{\"dude\":" +
								stamp + "}}]}}"
					}),
					GenCallBack(func(se *Session, st int, storage []byte) {
						ss._hex_cache[x][y] <- v + 1
					})
			}),
			GenSession(func() (float32, GenCall, GenCallBack) {
				return 50,
					GenCall(func(ps ...string) (_m, _t, _u, _b string) {

						return "GET", "REST",
							"http://localhost:58888/v1/objs/" + strconv.Itoa(x) +
								"," + strconv.Itoa(y),
							""
					}),
					nil
			}),
		})
	}
}

func (ss *WorldServerScenario) NextCall(rg *rand.Rand) (*Call, error) {
	for {
		i := rg.Intn(ss.SessionAmount)
		if i < 0 || i >= ss.SessionAmount {
			log.Println("i")
		}
		select {
		case st := <-ss._sessions[i].StepLock:
			switch st {
			case STEP1:
				if ss._sessions[i]._calls[st].GenParam != nil {
					ss._sessions[i]._calls[st].Method, ss._sessions[i]._calls[st].Type, ss._sessions[i]._calls[st].URL, ss._sessions[i]._calls[st].Body = ss._sessions[i]._calls[st].GenParam()
				}
				// execute session call for the first time
				return ss._sessions[i]._calls[st], nil
			default:
				// choose a non-initialized call randomly
				ss._sessions[i].StepLock <- REST
				q := rg.Float32() * ss._sessions[i]._totalWeight
				for j := STEP1 + 1; j < ss._sessions[i]._count; j++ {
					if q <= ss._sessions[i]._calls[j].RandomWeight {
						// add 1 to seq
						if ss._sessions[i]._calls[j].GenParam != nil {
							ss._sessions[i]._calls[j].Method, ss._sessions[i]._calls[j].Type, ss._sessions[i]._calls[j].URL, ss._sessions[i]._calls[j].Body = ss._sessions[i]._calls[j].GenParam()
						}
						return ss._sessions[i]._calls[j], nil
					}
				}
			}
		default:
			continue
		}

	}

	log.Fatal("what? should never reach here")
	return nil, errors.New("all sessions are being initialized")
}

func (s *WorldServerScenario) CustomizedReport() string {
	return ""
}

func init() {
	Register("ws_session", newWorldServerScenario)
}

func newWorldServerScenario(size int) (Profile, error) {
	return &WorldServerScenario{
		SessionAmount: size,
	}, nil
}

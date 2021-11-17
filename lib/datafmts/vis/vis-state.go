package datafmts

// data types for the vis part.

type QueryRunnerTable struct {
  Queries []QueryRow
}

type QueryRow struct {
  Key         string
  Type        string
  RunnerState QueryRunnerState
}

type QueryRunnerState struct {
  PeersSeen      int
  PeersQueried   int
  PeersDialed    int
  PeersToQuery   int
  PeersRemaining int
  RateLimit      RateLimit
  Result         QueryResult
  StartTime      string
  CurrTime       string
  EndTime        string
}

type QueryResult struct {
  Success     bool
  FoundPeer   string
  CloserPeers int
  FinalSet    int
  QueriedSet  int
}

type QueryPeerRow struct {
  QueryID       string
  QueryOrder    int // sequentially counting up rows.
  PeerID        string
  XORDistance   int
  Hops          int
  Spans         []Span
  TotalDuration string

  // results data
  Records int // if any records received

  CloserPeersRecv int // total new peers received
  CloserPeersNew  int // number of peers which were new

  ProviderPeersRecv int // total new peers received
  ProviderPeersNew  int // number of peers which were new
}

type Span struct {
  Type     string // Dial, Request, etc.
  Start    string // timestamp
  End      string // timestamp
  Duration string // time duration
}

type RateLimit struct {
  Capacity int
  Length   int
}

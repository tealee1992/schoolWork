package monitor
// Copyright 2013 Google Inc. All Rights Reserved.
2 //
3 // Licensed under the Apache License, Version 2.0 (the "License");
4 // you may not use this file except in compliance with the License.
5 // You may obtain a copy of the License at
6 //
7 //     http://www.apache.org/licenses/LICENSE-2.0
8 //
9 // Unless required by applicable law or agreed to in writing, software
10 // distributed under the License is distributed on an "AS IS" BASIS,
11 // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
12 // See the License for the specific language governing permissions and
13 // limitations under the License.
14
15 package daemon
16
17 import (
18         "errors"
19         "flag"
20         "fmt"
21         "net"
22         "os"
23         "strconv"
24         "strings"
25         "sync"
26 )
27
28 // ErrStopped is returned when Accept is called on a listener
29 // which has been stopped.
30 var ErrStopped = errors.New("daemon: listener stopped")
31
32 // ErrTimeout is returned when Restart times out.
33 var ErrTimeout = errors.New("daemon: timeout")
34
35 type waitConn struct {
36         *sync.WaitGroup
37         net.Conn
38         closeOnce sync.Once
39 }
40
41 func (c *waitConn) Close() error {
	42         err := fmt.Errorf("double close")
	43         c.closeOnce.Do(func() {
		44                 defer c.Done()
		45                 Verbose.Printf("Closed connection: (local) %s <- %s (remote)",
			46                         c.LocalAddr(), c.RemoteAddr())
		47                 err = c.Conn.Close()
		48         })
	49         return err
	50 }
51
52 // A WaitListener is a listener which accepts connections like a normal
53 // Listener, but counts them and can Wait for all of them to close.
54 type WaitListener struct {
55         wg sync.WaitGroup
56         net.Listener
57         stop chan bool
58 }
59
60 // Accept is a wrapper around the underlying Listener's accept
61 // to facilitate tracking connections.
62 func (w *WaitListener) Accept() (conn net.Conn, err error) {
	63         // To prevent race conditions, always assume we're going
	64         // to accept a connection.
	65         w.wg.Add(1)
	66         defer func() {
		67                 // If we didn't accept, decrement the count ourselves
		68                 if conn == nil {
			69                         w.wg.Done()
			70                 }
		71         }()
	72
	73         select {
	74         case <-w.stop:
75                 return nil, ErrStopped
76         default:
77         }
78
79         conn, err = w.Listener.Accept()
80         if err != nil {
81                 if strings.Contains(err.Error(), "closed network connection") {
82                         return nil, ErrStopped
83                 }
84                 return nil, err
85         }
86
87         Verbose.Printf("Accepted connection: (local) %s <- %s (remote)",
88                 conn.LocalAddr(), conn.RemoteAddr())
89
90         return &waitConn{
91                 WaitGroup: &w.wg,
92                 Conn:      conn,
93         }, nil
94 }
95
96 // Close stops and closes the listener; it is an error to close more than once.
97 func (w *WaitListener) Close() error {
	98         select {
	99         case <-w.stop:
100                 return fmt.Errorf("listener already closed")
101         default:
102                 close(w.stop)
103
104                 Verbose.Printf("Closing listener: %s", w.Addr())
105                 return w.Listener.Close()
106         }
107 }
108
109 // Stop stops the listener so that it can be used in another process.  After
110 // Stop, it may be necessary to create a dummy connection to this Listener to
111 // fall out of an existing Accept.  It is an error to call Stop more than once.
112 func (w *WaitListener) Stop() {
	113         close(w.stop)
	114
	115         Verbose.Printf("Stopping listener: %s", w.Addr())
	116 }
117
118 // File copies and the listener's underlying file descriptor.  This is intended
119 // to be used to pass the file descriptor on to a restarted version of this
120 // process.
121 func (w *WaitListener) File() *os.File {
	122         tcp, ok := w.Listener.(*net.TCPListener)
	123         if !ok {
		124                 Fatal.Printf("unknown listener type: %T", w.Listener)
		125         }
	126
	127         lf, err := tcp.File()
	128         if err != nil {
		129                 Fatal.Printf("failed to get fd: %s", err)
		130         }
	131         return lf
	132 }
133
134 // Wait waits for all associated connections to close.
135 func (w *WaitListener) Wait() {
	136         w.wg.Wait()
	137 }
138
139 // noop makes a dummy connection to the listener
140 func (w *WaitListener) noop() {
	141         addr := w.Addr().(*net.TCPAddr)
	142         for _, ip := range []net.IP{
		143                 net.IPv4(127, 0, 0, 1),
		144                 net.IPv6loopback,
		145                 addr.IP,
		146         } {
		147                 addr.IP = ip
		148                 conn, err := net.DialTCP("tcp", nil, addr)
		149                 if err != nil {
			150                         Verbose.Printf("noop(%q): %s", addr, err)
			151                         continue
			152                 }
		153                 defer conn.Close()
		154                 Verbose.Printf("noop(%q): Success", addr)
		155                 return
		156         }
	157         Verbose.Printf("noop(%q): failed to ping", addr)
	158 }
159
160 // A Listenable is something which can listen.  It can either
161 // be backed by a file descriptor of an existing listener,
162 // or if none is available, a new listener.  String returns
163 // the intended address for the listening socket as a string.
164 type Listenable interface {
165         Listen() (net.Listener, error)
166         String() string
167 }
168
169 type listenFlag struct {
170         flag, proto string
171         mode        string // "fd", "tcp"
172
173         // mode == "fd"
174         fd       int
175         listener *WaitListener
176
177         // mode == "tcp"
178         net   string
179         laddr *net.TCPAddr
180 }
181
182 func (l *listenFlag) Listen() (net.Listener, error) {
	183         var under net.Listener
	184         var err error
	185         switch l.mode {
		186         case "fd":
		187                 f := os.NewFile(uintptr(l.fd), fmt.Sprintf("&%d", l.fd))
		188                 under, err = net.FileListener(f)
		189         case "tcp":
		190                 under, err = net.ListenTCP(l.net, l.laddr)
		191         default:
		192                 return nil, fmt.Errorf("unknown mode %q", l.mode)
		193         }
	194         if err != nil {
		195                 return nil, err
		196         }
	197         Verbose.Printf("Listening for %s on: %s (from %s)", l.proto, under.Addr(), l.mode)
	198         listener := &WaitListener{
		199                 Listener: under,
		200                 stop:     make(chan bool),
		201         }
	202         l.listener = listener
	203         return listener, nil
	204 }
205
206 func (l *listenFlag) String() string {
	207         if l.laddr.IP == nil {
		208                 return fmt.Sprintf(":%d", l.laddr.Port)
		209         }
	210         return l.laddr.String()
	211 }
212
213 func (l *listenFlag) Set(s string) error {
	214         if len(s) == 0 {
		215                 return fmt.Errorf("--%s requires an argument", l.flag)
		216         }
	217
	218         // Check for passed file descriptor
	219         if s[0] == '&' {
		220                 fd, err := strconv.Atoi(s[1:])
		221                 if err != nil {
			222                         return fmt.Errorf("failed to parse &fd: %s", err)
			223                 }
		224                 l.mode, l.fd = "fd", fd
		225                 return nil
		226         }
	227
	228         laddr, err := net.ResolveTCPAddr(l.net, s)
	229         if err != nil {
		230                 return fmt.Errorf("failed to resolve %q: %s", s, err)
		231         }
	232         l.mode, l.laddr = "tcp", laddr
	233         return nil
	234 }
235
236 // ListenFlag registers a flag, which, when set, causes the returned
237 // Listenable to listen on the provided address.  If the flag is not
238 // provided, the default addr will be used.  The given proto is used
239 // to create the help text.
240 func ListenFlag(name, netw, addr, proto string) Listenable {
	241         laddr, err := net.ResolveTCPAddr(netw, addr)
	242         if err != nil {
		243                 Fatal.Printf("failed to resolve default %q: %s", addr, err)
		244         }
	245
	246         f := &listenFlag{
		247                 flag:  name,
		248                 proto: proto,
		249                 mode:  "tcp",
		250                 net:   netw,
		251                 laddr: laddr,
		252         }
	253         flag.Var(f, name, fmt.Sprintf("Address on which to listen for %s", proto))
	254         return f
	255 }
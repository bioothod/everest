package controllers

import "github.com/robfig/revel"
import "strings"
import "github.com/bioothod/elliptics-go/elliptics"
import "unsafe"
import "log"
import "C"
import "strconv"

var enode *elliptics.Node
var egroups []int

type App struct {
	*revel.Controller
}

func GoLogFunc(priv unsafe.Pointer, level int, msg *C.char) {
	var log *log.Logger = revel.TRACE

	switch level {
	case elliptics.ERROR:
		log = revel.ERROR
	case elliptics.INFO:
		log = revel.WARN
	case elliptics.NOTICE:
		log = revel.INFO
	case elliptics.DEBUG:
		log = revel.TRACE
	}

	log.Printf("%d: %s", level, C.GoString(msg))
}

var GoLogVar = GoLogFunc

func Init() {
	remotes, found := revel.Config.String("elliptics.remotes")
	if (!found) {
		revel.ERROR.Printf("no remote elliptics node")
		return
	}

	groups, found := revel.Config.String("elliptics.groups")
	if (!found) {
		revel.ERROR.Printf("no elliptics groups")
		return
	}

	for _, x := range strings.Split(groups, ":") {
		gint, err := strconv.ParseInt(x, 0, 0)
		if (err == nil) {
			egroups = append(egroups, int(gint))
		}
	}

	if (len(egroups) == 0) {
		revel.ERROR.Printf("no valid elliptics groups '%s'", groups)
		return
	}

	level, found := revel.Config.Int("elliptics.loglevel")
	if (!found) {
		level = 2
	}

	revel.ERROR.SetFlags(log.LstdFlags | log.Lmicroseconds)
	revel.WARN.SetFlags(log.LstdFlags | log.Lmicroseconds)
	revel.INFO.SetFlags(log.LstdFlags | log.Lmicroseconds)
	revel.TRACE.SetFlags(log.LstdFlags | log.Lmicroseconds)

	revel.INFO.Printf("controllers/app/Init: remotes: %s, groups: %s, log-level: %d\n", remotes, egroups, level)

	enode, err := elliptics.NewNodeLog(unsafe.Pointer(&GoLogVar), unsafe.Pointer(&GoLogVar), level)
	if err != nil {
		revel.ERROR.Println("failed to create new node: ", err)
		return
	}

	for _, x := range strings.Split(remotes, " ") {
		if err := enode.AddRemote(x); err != nil {
			revel.ERROR.Printf("could not add remote node: %s, err: %d\n", x, err)
		}
	}
	//defer enode.Free()
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) Hello(name string) revel.Result {
	c.Validation.Required(name).Message("Your name is required!")
	c.Validation.MinSize(name, 3).Message("Your name is not long enough!")

	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(App.Index)
	}

	return c.Render(name)
}

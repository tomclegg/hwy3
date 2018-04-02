var MP3Dir = {
    seconds: function(index, start, end) {
        var s = 0
        MP3Dir.intersect(index, start, end).map(function(interval) {
            s += interval[1]/1000
        })
        return s
    },
    intersect: function(index, starttime, endtime) {
        var start = starttime.getTime()
        var end = endtime.getTime()
        var intersect = []
        index.intervals.forEach(function(interval) {
            var istart = interval[0] * 1000
            var isec = interval[1] * 1000
            if (istart > end) return
            if (istart + isec < start) return
            if (istart + isec > end)
                isec = end - istart
            if (start > istart) {
                isec -= (start - istart)
                istart = start
            }
            intersect.push([istart, isec])
        })
        return intersect
    },
}
function toMetricDate(t) {
    return [t.getFullYear(), t.getMonth()+1, t.getDate()].map(function(i){return i.toString().padStart(2,'0')}).join('-')
}
function toMetricTime(t) {
    return [t.getHours(), t.getMinutes(), t.getSeconds()].map(function(i){return i.toString().padStart(2,'0')}).join(':')
}
function fromMetricDateTime(ymd, hms) {
    ymd = ymd.split('-')
    var t = new Date(parseInt(ymd[0]), parseInt(ymd[1])-1, parseInt(ymd[2]))
    var pm = hms.toUpperCase().indexOf('P') >= 0
    hms = hms.split(':')
    if (pm && hms[0]<12)
        hms[0] += 12
    t.setHours(parseInt(hms[0]))
    t.setMinutes(parseInt(hms[1]))
    if (hms[2])
        t.setSeconds(parseInt(hms[2]))
    return t
}
function toDisplayDuration(seconds) {
    var s = seconds % 60
    var m = Math.floor(seconds/60) % 60
    var h = Math.floor(seconds/3600)
    var dd = ''
    if (h>0) dd+=h+'h'
    if (m>0 || (h>0 && s>0)) dd+=m+'m'
    if (seconds<60 || s>0) dd+=s+'s'
    return dd
}
function toDisplaySize(bytes) {
    if (bytes>=1000000000) return ''+Math.floor(bytes/1000000000)+' GB'
    if (bytes>=1000000) return ''+Math.floor(bytes/1000000)+' MB'
    if (bytes>=1000) return ''+Math.floor(bytes/1000)+' KB'
    return ''+bytes+' byte'+(bytes==1?'':'s')
}
var MDC = {
    create: function(cls, vnode) {
        vnode.state.mdcComponent = new cls(vnode.dom)
    },
    remove: function(vnode) {
        vnode.state.mdcComponent.destroy()
    },
}
var Button = {
    oncreate: MDC.create.bind(null, mdc.ripple.MDCRipple),
    onremove: MDC.remove,
    view: function(vnode) {
        return m('button.mdc-button' + (vnode.attrs.raised ? '.mdc-button--raised' : ''), vnode.attrs, [
            m('i.material-icons.mdc-button__icon', vnode.attrs.icon),
            vnode.attrs.label,
        ])
    },
}
var TextField = {
    oncreate: MDC.create.bind(null, mdc.textField.MDCTextField),
    onremove: MDC.remove,
    view: function(vnode) {
        return m('.mdc-text-field.mdc-text-field--box.mdc-text-field--with-leading-icon', [
            m('i.material-icons.mdc-text-field__icon[tabindex=-1]', vnode.attrs.icon),
            m('input.mdc-text-field__input[type=text]', {
                id: vnode.attrs.id,
                value: vnode.attrs.store(),
                oninput: m.withAttr('value', vnode.attrs.store),
            }),
            m('label.mdc-floating-label[for=startdate]', vnode.attrs.label),
            m('.mdc-text-field__bottom-line'),
        ])
    },
}
var Archive = {
    oninit: function(vnode) {
        vnode.state.srv = m.stream({intervals:[]})
        m.request('/test/index.json').then(vnode.state.srv)
    },
    view: function(vnode) {
        return m('div', 'foo')
    },
}
var useCurrent = {
    view: function(vnode) {
        return [
            m('span', {
                style: {marginLeft: '2em'}
            }, vnode.attrs.src() && [
                m(Button, {
                    label: ['cut here (', vnode.attrs.src(), ')'],
                    icon: 'content_cut',
                    onclick: function() {
                        vnode.attrs.dst(vnode.attrs.src())
                    },
                }),
            ]),
        ]
    },
}
var IntervalMap = {
    Days: function(intervals) {
        if (!intervals || intervals.length < 1)
            return 0
        var t0 = new Date(intervals[0][0]*1000)
        t0.setHours(0)
        t0.setMinutes(0)
        t0.setSeconds(0)
        t0.setMilliseconds(0)
        var now = new Date()
        return Math.ceil((now.getTime() - t0.getTime())/86400000)
    },
    view: function(vnode) {
        var intervals = vnode.attrs.index.intervals
        if (!intervals || intervals.length < 1)
            return null
        var t0 = new Date(intervals[0][0]*1000)
        t0.setHours(0)
        t0.setMinutes(0)
        t0.setSeconds(0)
        t0.setMilliseconds(0)
        var now = new Date()
        var rows = []
        for (; t0<now; t0.setDate(t0.getDate()+1))
            rows.push(t0.getTime())
        var xscale = vnode.attrs.width / 86400000
        var yscale = vnode.attrs.rowHeight
        var yoffset = vnode.attrs.height
        return m('svg', vnode.attrs, rows.map(function(row, y) {
            var start = new Date(row)
            var end = y+1>=rows.length ? now : new Date(rows[y+1])
            return MP3Dir.intersect(vnode.attrs.index, start, end).map(function(intvl) {
                var x1 = intvl[0] + intvl[1] - start.getTime()
                var x0 = intvl[0] - start.getTime()
                return m('polyline', {
                    stroke: '#66d',
                    fill: '#ddf',
                    'stroke-width': 1,
                    points: [
                        [x0*xscale, yoffset - y*yscale],
                        [x1*xscale, yoffset - y*yscale],
                        [x1*xscale, yoffset - (y+.8)*yscale],
                        [x0*xscale, yoffset - (y+.8)*yscale],
                        [x0*xscale, yoffset - y*yscale],
                    ],
                })
            })
        }))
    },
}
var ArchivePage = {
    oninit: function(vnode) {
        var t = Date.now()
        var def = new Date(t - 86400000 - (t % 3600000))
        vnode.state.index = m.stream({intervals:[]})
        vnode.state.src = m.stream('/test')
        vnode.state.src.map(function(src) {
            m.request(src+'/index.json').then(vnode.state.index)
        })
        vnode.state.startdate = m.stream(toMetricDate(def))
        vnode.state.starttime = m.stream(toMetricTime(def))
        vnode.state.endtime = m.stream(toMetricTime(new Date(def.getTime() + 1800000)))
        vnode.state.want = m.stream.combine(function(index, startdate, starttime, endtime) {
            var okdate = /^ *[0-9]+-[0-9]+-[0-9]+ *$/
                if (!okdate.test(startdate()))
                    return {error: 'no date: '+startdate()}
            var oktime = /^ *[0-9]+:[0-9]+(:[0-9]+)? *([aApP][mM]?)? *$/
                if (!oktime.test(starttime()))
                    return {error: 'no time: '+starttime()}
            if (!oktime.test(endtime()))
                return {error: 'no time: '+endtime()}
            var start = fromMetricDateTime(startdate(), starttime())
            var end = fromMetricDateTime(startdate(), endtime())
            if (end < start)
                end.setDate(end.getDate()+1)
            if (end <= start)
                return {error: 'negative interval?'}
            var intvls = index().intervals
            if (intvls.length < 1 ||
                start.getTime()/1000 < intvls[0][0] ||
                end.getTime()/1000 > intvls[intvls.length-1][0] + intvls[intvls.length-1][1])
                return {error: 'data not available'}
            var seconds = MP3Dir.seconds(index(), start, end)
            if (seconds < 1)
                return {error: 'data not available'}
            var duration = toDisplayDuration(seconds)
            var filename = toMetricDate(start)+'_'+toMetricTime(start).replace(/:/g, '.')+'--'+duration+'.mp3'
            var size = toDisplaySize(seconds*index().bitRate/8)
            return {
                start: start,
                seconds: seconds,
                displayDuration: duration,
                displaySize: size,
                url: vnode.state.src() + '/' + Math.floor(start.getTime()/1000) + '-' + Math.floor(end.getTime()/1000) + '.mp3?filename='+filename,
            }
        }, [vnode.state.index, vnode.state.startdate, vnode.state.starttime, vnode.state.endtime])
        vnode.state.audioNode = m.stream(null)
        vnode.state.playerOffset = m.stream(null)
        vnode.state.ontimeupdate = function() {
            if (this !== vnode.state.audioNode())
                return
            var pos = this.currentTime
            if (pos > 0)
                pos = Math.round(pos)
            if (pos === vnode.state.playerOffset())
                return
            vnode.state.playerOffset(pos)
            m.redraw()
        }
        vnode.state.playerTime = m.stream.combine(function(want, offset) {
            if (!want().start)
                return null
            return toMetricTime(new Date(want().start.getTime() + 1000*offset()))
        }, [vnode.state.want, vnode.state.playerOffset])
        vnode.state.iframe = m.stream({})
    },
    view: function(vnode) {
        return m(Layout, [
            m('.mdc-layout-grid', [
                m('.mdc-layout-grid__inner', [
                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-12', [
                        m(TextField, {
                            id: 'startdate',
                            label: 'start date',
                            icon: 'event',
                            store: vnode.state.startdate,
                        }),
                    ]),
                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-12', [
                        m(TextField, {
                            id: 'starttime',
                            label: 'start time',
                            icon: 'timer',
                            store: vnode.state.starttime,
                        }),
                        vnode.state.playerOffset()>0 && vnode.state.want().url && m(useCurrent, {
                            dst: vnode.state.starttime,
                            src: vnode.state.playerTime,
                        }),
                    ]),
                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-12', [
                        m(TextField, {
                            id: 'endtime',
                            label: 'end time',
                            icon: 'timer',
                            store: vnode.state.endtime,
                        }),
                        vnode.state.playerOffset()>0 && vnode.state.want().url && m(useCurrent, {
                            dst: vnode.state.endtime,
                            src: vnode.state.playerTime,
                        }),
                    ]),
                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-8', [
                        m('audio', {
                            oncreate: function(audioNode) {
                                vnode.state.playerOffset(null)
                                vnode.state.audioNode(audioNode.dom)
                            },
                            ondurationchange: vnode.state.ontimeupdate,
                            onemptied: vnode.state.ontimeupdate,
                            onended: vnode.state.ontimeupdate,
                            onabort: vnode.state.ontimeupdate,
                            ontimeupdate: vnode.state.ontimeupdate,
                            key: vnode.state.want().url,
                            controls: true,
                            controlsList: 'nodownload',
                            preload: 'none',
                            style: {
                                width: '100%',
                            },
                        }, [
                            vnode.state.want().url && m('source', {
                                src: vnode.state.want().url,
                            }),
                        ]),
                    ]),
                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-12', [
                        m(Button, {
                            disabled: !vnode.state.want().url,
                            raised: true,
                            label: 'download',
                            icon: 'file_download',
                            onclick: function() {
                                vnode.state.iframe().src = vnode.state.want().url
                            },
                        }),
                        !vnode.state.want().seconds ? null : m('span', {style: {marginLeft: '2em'}}, [
                            vnode.state.want().displayDuration,
                            m.trust(' &mdash; '),
                            vnode.state.want().displaySize,
                        ]),
                    ]),
                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-12', [
                        m(TextField, {
                            id: 'src-uri',
                            label: 'stream source',
                            icon: 'audiotrack',
                            store: vnode.state.src,
                        }),
                    ]),
                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-12', [
                        m(IntervalMap, {
                            index: vnode.state.index(),
                            width: 600,
                            height: 20 * IntervalMap.Days(vnode.state.index().intervals),
                            rowHeight: 20,
                        }),
                    ]),
                ]),
            ]),
            m('iframe[width=0][height=0]', {
                oncreate: function(v) {
                    vnode.state.iframe(v.dom)
                },
            }),
        ])
    },
}
var Layout = {
    view: function(vnode) {
	return [
	    m('header.mdc-toolbar', [
                m('.mdc-toolbar__row', [
                    m('section.mdc-toolbar__section.mdc-toolbar__section--align-start', [
		        m('span.mdc-toolbar__title', 'archive'),
                    ]),
                    m('section.mdc-toolbar__section.mdc-toolbar__section--align-end', [
		        m('a.mdc-toolbar__menu-icon[href=/]', {oncreate: m.route.link}, 'recent'),
                    ]),
		]),
	    ]),
	    vnode.children,
	]
    },
}
m.route(document.body, "/", {
    "/": ArchivePage,
})

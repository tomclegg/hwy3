var DatePicker = {
    Weekdays: 'SMTWTFS'.split(''),
    Row: {
        view: function(vnode) {
            return m('div.mdc-typography--caption', {style: {width: '238px', margin: 'auto'}}, vnode.children.map(function(cell) {
                return m('span', {
                    style: {
                        display: 'inline-block',
                        position: 'relative',
                        width: '14.2857%',
                        lineHeight: '34px',
                        textAlign: 'center',
                    },
                }, cell)
            }))
        },
    },
    oninit: function(vnode) {
        vnode.state.tShowing = new Date(fromMetricDateTime(vnode.attrs.date(), '0:12:34'))
    },
    view: function(vnode) {
        var tSelected = fromMetricDateTime(vnode.attrs.date(), '0:12:34')
        var offerPrevMonth = toMetricDate(vnode.state.tShowing).slice(0, 7) > vnode.attrs.daysAvailable.first.slice(0, 7)
        var offerNextMonth = toMetricDate(vnode.state.tShowing).slice(0, 7) < vnode.attrs.daysAvailable.last.slice(0, 7)
        var t = new Date(vnode.state.tShowing)
        t.setDate(1)
        while (t.getDay() > 0)
            t.setDate(t.getDate()-1)
        return m('.mdc-card', {style: {userSelect: 'none'}}, [
            m('.', {style: {background: '#eee', padding: '.5em 1em', marginBottom: '1em'}}, [
                m('.mdc-typography--caption', {style: {float: 'right'}}, [
                    tSelected.toLocaleString('en', {year: 'numeric'}),
                ]),
                m('.mdc-typography--title', [
                    tSelected.toLocaleString('en', {weekday: 'short', month: 'short', day: 'numeric'}),
                ]),
            ]),
            m('.mdc-typography--body2', {style: {width: '238px', margin: 'auto'}}, [
                m('div', {
                    onclick: function() {
                        if (offerPrevMonth)
                            vnode.state.tShowing.setDate(0)
                    },
                    style: {display: 'inline-block', width: '14.2857%', textAlign: 'center', cursor: offerPrevMonth && 'pointer'},
                }, [
                    offerPrevMonth && m('i.material-icons', 'keyboard_arrow_left'),
                ]),
                m('div', {
                    style: {display: 'inline-block', width: '71.4285%', textAlign: 'center', verticalAlign: 'top'},
                }, [
                    vnode.state.tShowing.toLocaleString('en', {month: 'long', year: 'numeric'}),
                ]),
                m('div', {
                    onclick: function() {
                        if (offerNextMonth)
                            vnode.state.tShowing.setDate(32)
                    },
                    style: {display: 'inline-block', width: '14.2857%', textAlign: 'center', cursor: offerNextMonth && 'pointer'},
                }, [
                    offerNextMonth && m('i.material-icons', 'keyboard_arrow_right'),
                ]),
            ]),
            m('.', {style: {color: '#aaa'}}, [
                m(DatePicker.Row, DatePicker.Weekdays),
            ]),
            [0, 1, 2, 3, 4, 5].map(function() {
                // max calendar weeks in any month = 6
                var skip = true
                var cells = DatePicker.Weekdays.map(function() {
                    var month = t.getMonth()
                    var monthday = t.getDate()
                    var thismonth = month == vnode.state.tShowing.getMonth()
                    var selected = month == tSelected.getMonth() && monthday == tSelected.getDate()
                    var today = month == (new Date()).getMonth() && monthday == (new Date()).getDate()
                    var md = toMetricDate(t)
                    var quality = vnode.attrs.daysAvailable[md]
                    var selectable = quality > 0.01
                    t.setDate(monthday+1)
                    if (thismonth)
                        skip = false
                    return m('.', {
                        onclick: function() {
                            if (selectable)
                                vnode.attrs.date(md)
                        },
                        style: {
                            cursor: selectable ? 'pointer' : 'default',
                            width: '100%',
                            height: '100%',
                        },
                    }, [
                        selectable && m('.', {
                            style: {
                                backgroundColor: quality > 0.999 ? '#afc' : '#fb6',
                                position: 'absolute',
                                left: 0,
                                top: '10%',
                                width: '100%',
                                height: '80%',
                            },
                        }),
                        today && m('.', {
                            style: {
                                backgroundColor: '#fff',
                                position: 'absolute',
                                left: '10%',
                                top: '10%',
                                borderRadius: '50%',
                                width: '80%',
                                height: '80%',
                            },
                        }),
                        m('.', {
                            style: {
                                backgroundColor: '#6200ee',
                                position: 'absolute',
                                left: 0,
                                top: 0,
                                borderRadius: '50%',
                                width: '100%',
                                height: '100%',
                                transform: 'scale('+(selected ? 1 : 0)+')',
                                transition: 'all 450ms cubic-bezier(0.2, 1, 0.3, 1) 0ms',
                            },
                        }),
                        m('span', {
                            style: {
                                color: selected ? '#fff' : today ? '#6200ee' : thismonth ? '#000' : '#aaa',
                                opacity: 0.999,
                                fontWeight: selected ? 'bold' : 'normal',
                            }
                        }, [
                            monthday,
                        ]),
                    ])
                })
                return m(DatePicker.Row, skip ? [m.trust('&nbsp;')] : cells)
            }),
        ])
    },
}
var MP3Dir = {
    // number of seconds available between start and end
    seconds: function(index, start, end) {
        var s = 0
        MP3Dir.intersect(index, start, end).map(function(interval) {
            s += interval[1]/1000
        })
        return s
    },
    // map of metricdate -> seconds available, 'first' -> earliest
    // metricdate, and 'last' -> latest metricdate
    daysAvailable: function(index) {
        var max = new Date()
        var lastday = toMetricDate(max)
        var min = max
        if (index.intervals.length > 0)
            min = new Date(index.intervals[0][0] * 1000)
        var t = new Date(min)
        var day = toMetricDate(t)
        var days = {first: day}
        var nextday, istart, iend
        while (day <= lastday) {
            t.setDate(t.getDate()+1)
            nextday = toMetricDate(t)
            istart = fromMetricDateTime(day, '0:00')
            iend = fromMetricDateTime(nextday, '0:00')
            days[day] = MP3Dir.seconds(index, istart, iend) * 1000 / (iend.getTime() - istart.getTime())
            day = nextday
            days.last = day
        }
        return days
    },
    intersect: function(index, start, end) {
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
        return vnode.state.mdcComponent
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
        return m('.mdc-text-field.mdc-text-field--box.mdc-text-field--with-leading-icon', vnode.attrs, [
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
var useCurrent = {
    view: function(vnode) {
        var t = vnode.attrs.src() && toMetricTime(vnode.attrs.src())
        return m(Button, {
            disabled: !t || vnode.attrs.disabled,
            label: t ? ['cut here (', t, ')'] : 'cut here',
            icon: 'content_cut',
            onclick: function() {
                vnode.attrs.dst(t)
            },
        })
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
    oninit: function(vnode) {
        vnode.state.width = 0
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
        var yoffset = vnode.attrs.height - yscale/2
        return m('svg', vnode.attrs, rows.map(function(row, y) {
            var start = row
            var end = y+1>=rows.length ? now.getTime() : rows[y+1]
            var t0, t1 = start
            y = yoffset - y*yscale
            var selection = {
                color: '#00f',
                stroke: .5,
                strokeDashArray: '1, 1',
            }
            var available = {
                color: '#afc',
                border: true,
                stroke: .8,
            }
            var unavailable = {
                color: '#fb6',
                border: true,
                stroke: 1,
            }
            var hunk = function(hunktype, t0, t1) {
                var x0 = (t0 - start) * xscale
                var x1 = (t1 - start) * xscale
                return [
                    m('polyline', {
                        stroke: hunktype.color,
                        'stroke-width': hunktype.stroke * yscale,
                        'stroke-dasharray': hunktype.strokeDashArray,
                        points: [
                            [x0, y],
                            [x1, y],
                        ],
                    }),
                    hunktype.border && m('polyline', {
                        stroke: '#fff',
                        'stroke-width': yscale * .4,
                        points: [
                            [x0-1, y],
                            [x0, y],
                        ],
                    }),
                ]
            }
            var segments = MP3Dir.intersect(vnode.attrs.index, start, end)
            return [
                (segments.length>0 && segments[0][0]>start) && hunk(unavailable, start, segments[0][0]),
            ].concat(segments.map(function(seg, idx) {
                t0 = seg[0]
                t1 = seg[0] + seg[1]
                return [
                    hunk(available, t0, t1),
                    (idx<segments.length-1 && t1<segments[idx+1][0]) && hunk('#fc8', t1, segments[idx+1][0]),
                ]
            })).concat([
                t1<end && hunk(unavailable, t1, end),
                vnode.attrs.selection && hunk(selection, vnode.attrs.selection[0], vnode.attrs.selection[1]),
                m('text', {
                    x: vnode.attrs.width/100,
                    y: y+yscale*.2,
                    'font-size': yscale*.5,
                    'font-family': 'Roboto, sans-serif',
                }, [
                    new Date(start).toLocaleString('en', {month: 'long', day: 'numeric'}),
                    ': ',
                    new Date(t0 ? segments[0][0] : start).toLocaleString('en', {hour: 'numeric', minute: '2-digit'}).toLocaleLowerCase().replace(':00', ''),
                    ' - ',
                    new Date(t1).toLocaleString('en', {hour: 'numeric', minute: '2-digit'}).toLocaleLowerCase().replace(':00', ''),
                ]),
            ])
        }))
    },
}
var Clockface = {
    view: function(vnode) {
        var r = vnode.attrs.width/2
        return m('svg', vnode.attrs, [
            [1,2,3,4,5,6,7,8,9,10,11,12].map(function(hh) {
                return [1,2,3,4].map(function(mm) {
                    return m('polyline', {
                        stroke: '#aaa',
                        'stroke-width': 1,
                        points: [[r,0], [r,r/14]],
                        transform: 'rotate('+[30*hh+6*mm, [r,r]]+')',
                    })
                }).concat([
                    m('polyline', {
                        stroke: '#aaa',
                        'stroke-width': 4,
                        points: [[r,0], [r,r/7]],
                        transform: 'rotate('+[30*hh, [r,r]]+')',
                    }),
                ])
            }),
            vnode.attrs.time === null ? null : [
                m('polyline', {
                    stroke: '#000',
                    'stroke-width': 4,
                    points: [[r,r], [r,r/2]],
                    transform: 'rotate('+[30*(vnode.attrs.time.getHours()+vnode.attrs.time.getMinutes()/60), [r,r]]+')',
                }),
                m('polyline', {
                    stroke: '#000',
                    'stroke-width': 2,
                    points: [[r,r], [r,r/7]],
                    transform: 'rotate('+[6*(vnode.attrs.time.getMinutes()+vnode.attrs.time.getSeconds()/60), [r,r]]+')',
                }),
                m('polyline', {
                    stroke: '#b00',
                    points: [[r,r*8/7], [r,0]],
                    transform: 'rotate('+[6*(vnode.attrs.time.getSeconds()+vnode.attrs.time.getMilliseconds()/1000), [r,r]]+')',
                }),
                m('circle', {
                    fill: '#b00',
                    cx: r,
                    cy: r,
                    r: r/16,
                }),
            ],
        ])
    },
}
var ArchivePage = {
    oninit: function(vnode) {
        var t = Date.now()
        var def = new Date(t - 86400000 - (t % 3600000))
        vnode.state.index = m.stream({intervals:[]})
        m.request('/'+vnode.attrs.channel+'/index.json').then(vnode.state.index)
        vnode.state.daysAvailable = vnode.state.index.map(MP3Dir.daysAvailable)
        vnode.state.startdate = m.stream(vnode.attrs.startdate || toMetricDate(def))
        vnode.state.starttime = m.stream(vnode.attrs.starttime || toMetricTime(def))
        vnode.state.endtime = m.stream(vnode.attrs.endtime || toMetricTime(new Date(def.getTime() + 1800000)))
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
            var seconds = MP3Dir.seconds(index(), start.getTime(), end.getTime())
            if (seconds < 1)
                return {error: 'data not available'}
            var duration = toDisplayDuration(seconds)
            var filename = toMetricDate(start)+'_'+toMetricTime(start).replace(/:/g, '.')+'--'+duration+'.mp3'
            var size = toDisplaySize(seconds*index().bitRate/8)
            return {
                start: start,
                end: end,
                seconds: seconds,
                displayDuration: duration,
                displaySize: size,
                url: '/' + vnode.attrs.channel + '/' + Math.floor(start.getTime()/1000) + '-' + Math.floor(end.getTime()/1000) + '.mp3?filename='+filename,
            }
        }, [vnode.state.index, vnode.state.startdate, vnode.state.starttime, vnode.state.endtime])
        vnode.state.audioNode = m.stream(null)
        vnode.state.playerOffset = m.stream(null)
        vnode.state.playerResume = m.stream(false)
        vnode.state.ontimeupdate = function() {
            if (this !== vnode.state.audioNode())
                return
            var pos = this.currentTime
            if (pos === undefined)
                pos = null
            if (pos === vnode.state.playerOffset())
                return
            if (pos !== null && Math.floor(pos) == Math.floor(vnode.state.playerOffset()))
                return
            vnode.state.playerOffset(pos)
            m.redraw()
        }
        vnode.state.playerTime = m.stream.combine(function(want, offset) {
            if (!want().start)
                return null
            return new Date(want().start.getTime() + 1000*offset())
        }, [vnode.state.want, vnode.state.playerOffset])
        vnode.state.want.map(function() {
            m.route.set('/archive/:channel/:startdate/:starttime/:endtime', {
                channel: vnode.attrs.channel,
                startdate: vnode.state.startdate(),
                starttime: vnode.state.starttime(),
                endtime: vnode.state.endtime(),
            }, {
                replace: true,
            })
        })
        vnode.state.iframe = m.stream({})
    },
    view: function(vnode) {
        return m(Layout, [
            m('.mdc-layout-grid', [
                m('.mdc-layout-grid__inner', [
                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-4', [
                        m('div', {style: {width: '272px'}}, [
                            m(DatePicker, {
                                date: vnode.state.startdate,
                                daysAvailable: vnode.state.daysAvailable(),
                            }),
                        ]),
                    ]),
                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-8', [
                        m('.mdc-layout-grid__inner', [
                            m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-6', [
                                m(TextField, {
                                    id: 'starttime',
                                    label: 'start time',
                                    icon: 'timer',
                                    store: vnode.state.starttime,
                                    style: {marginTop: 0},
                                }),
                                m(useCurrent, {
                                    disabled: vnode.state.playerOffset()===null || !vnode.state.want().url,
                                    dst: function(t) {
                                        vnode.state.starttime(t)
                                        vnode.state.playerResume(true)
                                    },
                                    src: vnode.state.playerTime,
                                })
                            ]),
                            m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-6', [
                                m(TextField, {
                                    id: 'endtime',
                                    label: 'end time',
                                    icon: 'timer',
                                    store: vnode.state.endtime,
                                    style: {marginTop: 0},
                                }),
                                m(useCurrent, {
                                    disabled: vnode.state.playerOffset()===null || !vnode.state.want().url,
                                    dst: vnode.state.endtime,
                                    src: vnode.state.playerTime,
                                }),
                            ]),
                            m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-2', {style: {textAlign: 'right'}}, [
                                m(Clockface, {
                                    width: 100,
                                    height: 100,
                                    time: vnode.state.playerTime(),
                                }),
                            ]),
                            m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-10', [
                                m('.', {style: {marginBottom: '1em'}}, [
                                    m('audio', {
                                        oncreate: function(audioNode) {
                                            vnode.state.playerOffset(null)
                                            vnode.state.audioNode(audioNode.dom)
                                            audioNode.state.onkeydown = function(e) {
                                                if (document.activeElement.tagName == 'INPUT')
                                                    return
                                                else if (e.key === 'ArrowRight')
                                                    audioNode.dom.currentTime = Math.max(Math.floor((audioNode.dom.currentTime+5)/5)*5, 0)
                                                else if (e.key === 'ArrowLeft')
                                                    audioNode.dom.currentTime = Math.max(Math.floor((audioNode.dom.currentTime-1)/5)*5, 0)
                                                else if (e.key === ' ')
                                                    audioNode.dom.paused ? audioNode.dom.play() : audioNode.dom.pause()
                                                else
                                                    return
                                                e.preventDefault()
                                            }
                                            document.body.addEventListener('keydown', audioNode.state.onkeydown, {capture: true})
                                        },
                                        onremove: function(audioNode) {
                                            document.body.removeEventListener('keydown', audioNode.state.onkeydown, {capture: true})
                                        },
                                        ondurationchange: vnode.state.ontimeupdate,
                                        onemptied: vnode.state.ontimeupdate,
                                        onended: vnode.state.ontimeupdate,
                                        onabort: vnode.state.ontimeupdate,
                                        ontimeupdate: vnode.state.ontimeupdate,
                                        controls: true,
                                        controlsList: 'nodownload',
                                        preload: 'metadata',
                                        style: {
                                            width: '100%',
                                        },
                                    }, [
                                        vnode.state.want().url && m('source', {
                                            onupdate: function(vnode) {
                                                var audio = vnode.dom.parentElement
                                                if (vnode.state.url !== vnode.attrs.channel) {
                                                    vnode.state.url = vnode.attrs.channel
                                                    audio.autoplay = vnode.attrs.resume() && audio.buffered.length>0 && !audio.paused
                                                    audio.load()
                                                    vnode.attrs.resume(false)
                                                }
                                            },
                                            onbeforeremove: function(vnode) {
                                                vnode.state.parent = vnode.dom.parentElement
                                            },
                                            onremove: function(vnode) {
                                                vnode.state.parent.load()
                                            },
                                            src: vnode.state.want().url,
                                            resume: vnode.state.playerResume,
                                        }),
                                    ]),
                                ]), m('.', [
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
                            ]),
                        ]),
                    ]),
                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-8', {
                        oncreate: function(cell) {
                            vnode.state.mapWidth = cell.dom.getBoundingClientRect().width
                            m.redraw()
                        },
                        onupdate: function(cell) {
                            var cw = cell.dom.getBoundingClientRect().width
                            if (vnode.state.mapWidth === cw) return
                            vnode.state.mapWidth = cw
                            m.redraw()
                        },
                    }, [
                        m(IntervalMap, {
                            index: vnode.state.index(),
                            width: vnode.state.mapWidth,
                            height: 24 * IntervalMap.Days(vnode.state.index().intervals),
                            rowHeight: 24,
                            selection: vnode.state.want().start ? [vnode.state.want().start, vnode.state.want().end] : null,
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
    oninit: function(vnode) {
        vnode.state.drawer = null
        vnode.state.channels = m.stream({})
        m.request('/sys/channels').then(vnode.state.channels)
        vnode.state.theme = m.stream({})
        m.request('/sys/theme').then(vnode.state.theme)
    },
    view: function(vnode) {
	return [
	    m('header.mdc-top-app-bar', {
                oncreate: MDC.create.bind(null, mdc.topAppBar.MDCTopAppBar),
                onremove: MDC.remove,
            }, [
                m('.mdc-top-app-bar__row', [
                    m('section.mdc-top-app-bar__section.mdc-top-app-bar__section--align-start', [
                        m('a[href=#].material-icons.mdc-top-app-bar__navigation-icon', {
                            onclick: function() {
                                vnode.state.drawer.open = true
                                return false
                            },
                        }, 'menu'),
		        m('span.mdc-top-app-bar__title', vnode.state.theme().title || 'Audio archive'),
                    ]),
	        ]),
	    ]),
            m('aside.mdc-drawer.mdc-drawer--temporary.mdc-typography', {
                oncreate: function(drawernode) {
                    vnode.state.drawer = MDC.create(mdc.drawer.MDCTemporaryDrawer, drawernode)
                    if (m.route.get() === '/')
                        vnode.state.drawer.open = true
                },
                onremove: MDC.remove,
            }, [
                m('nav.mdc-drawer__drawer', [
                    m('header.mdc-drawer__header', [
                        m('.mdc-drawer__header-content', 'Available channels:'),
                    ]),
                    m('nav.mdc-drawer__content.mdc-list', [
                        Object.keys(vnode.state.channels()).sort().map(function(name) {
                            if (name.indexOf('/')!=0 || !vnode.state.channels()[name].archive)
                                return null
                            return m('a.mdc-list-item', {
                                className: m.route.param('channel')==name.slice(1) && 'mdc-list-item--activated',
                                href: '/archive'+name,
                                oncreate: m.route.link,
                            }, [
                                m('i.material-icons.mdc-list-item__graphic[aria-hidden=true]', 'music_note'),
                                name,
                            ])
                        }),
                    ]),
                ]),
            ]),
	    vnode.children,
	]
    },
}
var ArchiveRoute = {
    view: function(vnode) {
        return m(ArchivePage, {
            channel: m.route.param('channel'),
            startdate: m.route.param('startdate'),
            starttime: m.route.param('starttime'),
            endtime: m.route.param('endtime'),
        })
    }
}
m.route(document.body, "/", {
    "/": Layout,
    "/archive/:channel": ArchiveRoute,
    "/archive/:channel/:startdate/:starttime/:endtime": ArchiveRoute,
})
window.onresize = m.redraw

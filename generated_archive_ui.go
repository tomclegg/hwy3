package main

import "github.com/tomclegg/canfs"

var archiveUI = canfs.FileSystem{Content: map[string]canfs.FileInfo{
	"/archive.js": {N: "archive.js", M: 0x1a4, S: 33653, MT: 1525817099611952100, FileData: canfs.StringData{Data: "var DatePicker = {\n    Weekdays: 'SMTWTFS'.split(''),\n    Row: {\n        view: function(vnode) {\n            return m('div.mdc-typography--caption', {style: {width: '238px', margin: 'auto'}}, vnode.children.map(function(cell) {\n                return m('span', {\n                    style: {\n                        display: 'inline-block',\n                        position: 'relative',\n                        width: '14.2857%',\n                        lineHeight: '34px',\n                        textAlign: 'center',\n                    },\n                }, cell)\n            }))\n        },\n    },\n    oninit: function(vnode) {\n        vnode.state.tShowing = new Date(fromMetricDateTime(vnode.attrs.date(), '0:12:34'))\n    },\n    view: function(vnode) {\n        var tSelected = fromMetricDateTime(vnode.attrs.date(), '0:12:34')\n        var offerPrevMonth = toMetricDate(vnode.state.tShowing).slice(0, 7) > vnode.attrs.daysAvailable.first.slice(0, 7)\n        var offerNextMonth = toMetricDate(vnode.state.tShowing).slice(0, 7) < vnode.attrs.daysAvailable.last.slice(0, 7)\n        var t = new Date(vnode.state.tShowing)\n        t.setDate(1)\n        while (t.getDay() > 0)\n            t.setDate(t.getDate()-1)\n        return m('.mdc-card', {style: {userSelect: 'none'}}, [\n            m('.', {style: {background: '#eee', padding: '.5em 1em', marginBottom: '1em'}}, [\n                m('.mdc-typography--caption', {style: {float: 'right'}}, [\n                    tSelected.toLocaleString('en', {year: 'numeric'}),\n                ]),\n                m('.mdc-typography--title', [\n                    tSelected.toLocaleString('en', {weekday: 'short', month: 'short', day: 'numeric'}),\n                ]),\n            ]),\n            m('.mdc-typography--body2', {style: {width: '238px', margin: 'auto'}}, [\n                m('div', {\n                    onclick: function() {\n                        if (offerPrevMonth)\n                            vnode.state.tShowing.setDate(0)\n                    },\n                    style: {display: 'inline-block', width: '14.2857%', textAlign: 'center', cursor: offerPrevMonth && 'pointer'},\n                }, [\n                    offerPrevMonth && m('i.material-icons', 'keyboard_arrow_left'),\n                ]),\n                m('div', {\n                    style: {display: 'inline-block', width: '71.4285%', textAlign: 'center', verticalAlign: 'top'},\n                }, [\n                    vnode.state.tShowing.toLocaleString('en', {month: 'long', year: 'numeric'}),\n                ]),\n                m('div', {\n                    onclick: function() {\n                        if (offerNextMonth)\n                            vnode.state.tShowing.setDate(32)\n                    },\n                    style: {display: 'inline-block', width: '14.2857%', textAlign: 'center', cursor: offerNextMonth && 'pointer'},\n                }, [\n                    offerNextMonth && m('i.material-icons', 'keyboard_arrow_right'),\n                ]),\n            ]),\n            m('.', {style: {color: '#aaa'}}, [\n                m(DatePicker.Row, DatePicker.Weekdays),\n            ]),\n            [0, 1, 2, 3, 4, 5].map(function() {\n                // max calendar weeks in any month = 6\n                var skip = true\n                var cells = DatePicker.Weekdays.map(function() {\n                    var month = t.getMonth()\n                    var monthday = t.getDate()\n                    var thismonth = month == vnode.state.tShowing.getMonth()\n                    var selected = month == tSelected.getMonth() && monthday == tSelected.getDate()\n                    var today = month == (new Date()).getMonth() && monthday == (new Date()).getDate()\n                    var md = toMetricDate(t)\n                    var quality = vnode.attrs.daysAvailable[md]\n                    var selectable = quality > 0.01\n                    t.setDate(monthday+1)\n                    if (thismonth)\n                        skip = false\n                    return m('.', {\n                        onclick: function() {\n                            if (selectable)\n                                vnode.attrs.date(md)\n                        },\n                        style: {\n                            cursor: selectable ? 'pointer' : 'default',\n                            width: '100%',\n                            height: '100%',\n                        },\n                    }, [\n                        selectable && m('.', {\n                            style: {\n                                backgroundColor: quality > 0.999 ? '#afc' : '#fb6',\n                                position: 'absolute',\n                                left: 0,\n                                top: '10%',\n                                width: '100%',\n                                height: '80%',\n                            },\n                        }),\n                        today && m('.', {\n                            style: {\n                                backgroundColor: '#fff',\n                                position: 'absolute',\n                                left: '10%',\n                                top: '10%',\n                                borderRadius: '50%',\n                                width: '80%',\n                                height: '80%',\n                            },\n                        }),\n                        m('.', {\n                            style: {\n                                backgroundColor: '#6200ee',\n                                position: 'absolute',\n                                left: 0,\n                                top: 0,\n                                borderRadius: '50%',\n                                width: '100%',\n                                height: '100%',\n                                transform: 'scale('+(selected ? 1 : 0)+')',\n                                transition: 'all 450ms cubic-bezier(0.2, 1, 0.3, 1) 0ms',\n                            },\n                        }),\n                        m('span', {\n                            style: {\n                                color: selected ? '#fff' : today ? '#6200ee' : thismonth ? '#000' : '#aaa',\n                                opacity: 0.999,\n                                fontWeight: selected ? 'bold' : 'normal',\n                            }\n                        }, [\n                            monthday,\n                        ]),\n                    ])\n                })\n                return m(DatePicker.Row, skip ? [m.trust('&nbsp;')] : cells)\n            }),\n        ])\n    },\n}\nvar MP3Dir = {\n    // number of seconds available between start and end\n    seconds: function(index, start, end) {\n        var s = 0\n        MP3Dir.intersect(index, start, end).map(function(interval) {\n            s += interval[1]/1000\n        })\n        return s\n    },\n    // map of metricdate -> seconds available, 'first' -> earliest\n    // metricdate, and 'last' -> latest metricdate\n    daysAvailable: function(index) {\n        var max = new Date()\n        var lastday = toMetricDate(max)\n        var min = max\n        if (index.intervals.length > 0)\n            min = new Date(index.intervals[0][0] * 1000)\n        var t = new Date(min)\n        var day = toMetricDate(t)\n        var days = {first: day}\n        var nextday, istart, iend\n        while (day <= lastday) {\n            t.setDate(t.getDate()+1)\n            nextday = toMetricDate(t)\n            istart = fromMetricDateTime(day, '0:00')\n            iend = fromMetricDateTime(nextday, '0:00')\n            days[day] = MP3Dir.seconds(index, istart, iend) * 1000 / (iend.getTime() - istart.getTime())\n            day = nextday\n            days.last = day\n        }\n        return days\n    },\n    intersect: function(index, start, end) {\n        var intersect = []\n        index.intervals.forEach(function(interval) {\n            var istart = interval[0] * 1000\n            var isec = interval[1] * 1000\n            if (istart > end) return\n            if (istart + isec < start) return\n            if (istart + isec > end)\n                isec = end - istart\n            if (start > istart) {\n                isec -= (start - istart)\n                istart = start\n            }\n            intersect.push([istart, isec])\n        })\n        return intersect\n    },\n}\nfunction toMetricDate(t) {\n    return [t.getFullYear(), t.getMonth()+1, t.getDate()].map(function(i){return i.toString().padStart(2,'0')}).join('-')\n}\nfunction toMetricTime(t) {\n    return [t.getHours(), t.getMinutes(), t.getSeconds()].map(function(i){return i.toString().padStart(2,'0')}).join(':')\n}\nfunction fromMetricDateTime(ymd, hms) {\n    ymd = ymd.split('-')\n    var t = new Date(parseInt(ymd[0]), parseInt(ymd[1])-1, parseInt(ymd[2]))\n    var pm = hms.toUpperCase().indexOf('P') >= 0\n    hms = hms.split(':')\n    if (pm && hms[0]<12)\n        hms[0] += 12\n    t.setHours(parseInt(hms[0]))\n    t.setMinutes(parseInt(hms[1]))\n    if (hms[2])\n        t.setSeconds(parseInt(hms[2]))\n    return t\n}\nfunction toDisplayDuration(seconds) {\n    var s = seconds % 60\n    var m = Math.floor(seconds/60) % 60\n    var h = Math.floor(seconds/3600)\n    var dd = ''\n    if (h>0) dd+=h+'h'\n    if (m>0 || (h>0 && s>0)) dd+=m+'m'\n    if (seconds<60 || s>0) dd+=s+'s'\n    return dd\n}\nfunction toDisplaySize(bytes) {\n    if (bytes>=1000000000) return ''+Math.floor(bytes/1000000000)+' GB'\n    if (bytes>=1000000) return ''+Math.floor(bytes/1000000)+' MB'\n    if (bytes>=1000) return ''+Math.floor(bytes/1000)+' KB'\n    return ''+bytes+' byte'+(bytes==1?'':'s')\n}\nvar MDC = {\n    create: function(cls, vnode) {\n        vnode.state.mdcComponent = new cls(vnode.dom)\n        return vnode.state.mdcComponent\n    },\n    remove: function(vnode) {\n        vnode.state.mdcComponent.destroy()\n    },\n}\nvar Button = {\n    oncreate: MDC.create.bind(null, mdc.ripple.MDCRipple),\n    onremove: MDC.remove,\n    view: function(vnode) {\n        return m('button.mdc-button' + (vnode.attrs.raised ? '.mdc-button--raised' : ''), vnode.attrs, [\n            m('i.material-icons.mdc-button__icon', vnode.attrs.icon),\n            vnode.attrs.label,\n        ])\n    },\n}\nvar TextField = {\n    oncreate: MDC.create.bind(null, mdc.textField.MDCTextField),\n    onremove: MDC.remove,\n    view: function(vnode) {\n        return m('.mdc-text-field.mdc-text-field--box.mdc-text-field--with-leading-icon', vnode.attrs, [\n            m('i.material-icons.mdc-text-field__icon[tabindex=-1]', vnode.attrs.icon),\n            m('input.mdc-text-field__input[type=text]', {\n                id: vnode.attrs.id,\n                value: vnode.attrs.store(),\n                oninput: m.withAttr('value', vnode.attrs.store),\n            }),\n            m('label.mdc-floating-label[for=startdate]', vnode.attrs.label),\n            m('.mdc-text-field__bottom-line'),\n        ])\n    },\n}\nvar useCurrent = {\n    view: function(vnode) {\n        var t = vnode.attrs.src() && toMetricTime(vnode.attrs.src())\n        return m(Button, {\n            disabled: !t || vnode.attrs.disabled,\n            label: t ? ['cut here (', t, ')'] : 'cut here',\n            icon: 'content_cut',\n            onclick: function() {\n                vnode.attrs.dst(t)\n            },\n        })\n    },\n}\nvar IntervalMap = {\n    Days: function(intervals) {\n        if (!intervals || intervals.length < 1)\n            return 0\n        var t0 = new Date(intervals[0][0]*1000)\n        t0.setHours(0)\n        t0.setMinutes(0)\n        t0.setSeconds(0)\n        t0.setMilliseconds(0)\n        var now = new Date()\n        return Math.ceil((now.getTime() - t0.getTime())/86400000)\n    },\n    oninit: function(vnode) {\n        vnode.state.width = 0\n    },\n    view: function(vnode) {\n        var intervals = vnode.attrs.index.intervals\n        if (!intervals || intervals.length < 1)\n            return null\n        var t0 = new Date(intervals[0][0]*1000)\n        t0.setHours(0)\n        t0.setMinutes(0)\n        t0.setSeconds(0)\n        t0.setMilliseconds(0)\n        var now = new Date()\n        var rows = []\n        for (; t0<now; t0.setDate(t0.getDate()+1))\n            rows.push(t0.getTime())\n        var xscale = vnode.attrs.width / 86400000\n        var yscale = vnode.attrs.rowHeight\n        var yoffset = vnode.attrs.height - yscale/2\n        return m('svg', vnode.attrs, rows.map(function(row, y) {\n            var start = row\n            var end = y+1>=rows.length ? now.getTime() : rows[y+1]\n            var t0, t1 = start\n            y = yoffset - y*yscale\n            var selection = {\n                color: '#00f',\n                stroke: .5,\n                strokeDashArray: '1, 1',\n            }\n            var available = {\n                color: '#afc',\n                border: true,\n                stroke: .8,\n            }\n            var unavailable = {\n                color: '#fb6',\n                border: true,\n                stroke: 1,\n            }\n            var hunk = function(hunktype, t0, t1) {\n                var x0 = (t0 - start) * xscale\n                var x1 = (t1 - start) * xscale\n                return [\n                    m('polyline', {\n                        stroke: hunktype.color,\n                        'stroke-width': hunktype.stroke * yscale,\n                        'stroke-dasharray': hunktype.strokeDashArray,\n                        points: [\n                            [x0, y],\n                            [x1, y],\n                        ],\n                    }),\n                    hunktype.border && m('polyline', {\n                        stroke: '#fff',\n                        'stroke-width': yscale * .4,\n                        points: [\n                            [x0-1, y],\n                            [x0, y],\n                        ],\n                    }),\n                ]\n            }\n            var segments = MP3Dir.intersect(vnode.attrs.index, start, end)\n            return [\n                (segments.length>0 && segments[0][0]>start) && hunk(unavailable, start, segments[0][0]),\n            ].concat(segments.map(function(seg, idx) {\n                t0 = seg[0]\n                t1 = seg[0] + seg[1]\n                return [\n                    hunk(available, t0, t1),\n                    (idx<segments.length-1 && t1<segments[idx+1][0]) && hunk('#fc8', t1, segments[idx+1][0]),\n                ]\n            })).concat([\n                t1<end && hunk(unavailable, t1, end),\n                vnode.attrs.selection && hunk(selection, vnode.attrs.selection[0], vnode.attrs.selection[1]),\n                m('text', {\n                    x: vnode.attrs.width/100,\n                    y: y+yscale*.2,\n                    'font-size': yscale*.5,\n                    'font-family': 'Roboto, sans-serif',\n                }, [\n                    new Date(start).toLocaleString('en', {month: 'long', day: 'numeric'}),\n                    ': ',\n                    new Date(t0 ? segments[0][0] : start).toLocaleString('en', {hour: 'numeric', minute: '2-digit'}).toLocaleLowerCase().replace(':00', ''),\n                    ' - ',\n                    new Date(t1).toLocaleString('en', {hour: 'numeric', minute: '2-digit'}).toLocaleLowerCase().replace(':00', ''),\n                ]),\n            ])\n        }))\n    },\n}\nvar Clockface = {\n    view: function(vnode) {\n        var r = vnode.attrs.width/2\n        return m('svg', vnode.attrs, [\n            [1,2,3,4,5,6,7,8,9,10,11,12].map(function(hh) {\n                return [1,2,3,4].map(function(mm) {\n                    return m('polyline', {\n                        stroke: '#aaa',\n                        'stroke-width': 1,\n                        points: [[r,0], [r,r/14]],\n                        transform: 'rotate('+[30*hh+6*mm, [r,r]]+')',\n                    })\n                }).concat([\n                    m('polyline', {\n                        stroke: '#aaa',\n                        'stroke-width': 4,\n                        points: [[r,0], [r,r/7]],\n                        transform: 'rotate('+[30*hh, [r,r]]+')',\n                    }),\n                ])\n            }),\n            vnode.attrs.time === null ? null : [\n                m('polyline', {\n                    stroke: '#000',\n                    'stroke-width': 4,\n                    points: [[r,r], [r,r/2]],\n                    transform: 'rotate('+[30*(vnode.attrs.time.getHours()+vnode.attrs.time.getMinutes()/60), [r,r]]+')',\n                }),\n                m('polyline', {\n                    stroke: '#000',\n                    'stroke-width': 2,\n                    points: [[r,r], [r,r/7]],\n                    transform: 'rotate('+[6*(vnode.attrs.time.getMinutes()+vnode.attrs.time.getSeconds()/60), [r,r]]+')',\n                }),\n                m('polyline', {\n                    stroke: '#b00',\n                    points: [[r,r*8/7], [r,0]],\n                    transform: 'rotate('+[6*(vnode.attrs.time.getSeconds()+vnode.attrs.time.getMilliseconds()/1000), [r,r]]+')',\n                }),\n                m('circle', {\n                    fill: '#b00',\n                    cx: r,\n                    cy: r,\n                    r: r/16,\n                }),\n            ],\n        ])\n    },\n}\nvar ArchivePage = {\n    oninit: function(vnode) {\n        var t = Date.now()\n        var def = new Date(t - 86400000 - (t % 3600000))\n        vnode.state.index = m.stream({intervals:[]})\n        m.request('/'+vnode.attrs.channel+'/index.json').then(vnode.state.index)\n        vnode.state.daysAvailable = vnode.state.index.map(MP3Dir.daysAvailable)\n        vnode.state.startdate = m.stream(vnode.attrs.startdate || toMetricDate(def))\n        vnode.state.starttime = m.stream(vnode.attrs.starttime || toMetricTime(def))\n        vnode.state.endtime = m.stream(vnode.attrs.endtime || toMetricTime(new Date(def.getTime() + 1800000)))\n        vnode.state.want = m.stream.combine(function(index, startdate, starttime, endtime) {\n            var okdate = /^ *[0-9]+-[0-9]+-[0-9]+ *$/\n                if (!okdate.test(startdate()))\n                    return {error: 'no date: '+startdate()}\n            var oktime = /^ *[0-9]+:[0-9]+(:[0-9]+)? *([aApP][mM]?)? *$/\n                if (!oktime.test(starttime()))\n                    return {error: 'no time: '+starttime()}\n            if (!oktime.test(endtime()))\n                return {error: 'no time: '+endtime()}\n            var start = fromMetricDateTime(startdate(), starttime())\n            var end = fromMetricDateTime(startdate(), endtime())\n            if (end < start)\n                end.setDate(end.getDate()+1)\n            if (end <= start)\n                return {error: 'negative interval?'}\n            var intvls = index().intervals\n            if (intvls.length < 1 ||\n                start.getTime()/1000 < intvls[0][0] ||\n                end.getTime()/1000 > intvls[intvls.length-1][0] + intvls[intvls.length-1][1])\n                return {error: 'data not available'}\n            var seconds = MP3Dir.seconds(index(), start.getTime(), end.getTime())\n            if (seconds < 1)\n                return {error: 'data not available'}\n            var duration = toDisplayDuration(seconds)\n            var filename = toMetricDate(start)+'_'+toMetricTime(start).replace(/:/g, '.')+'--'+duration+'.mp3'\n            var size = toDisplaySize(seconds*index().bitRate/8)\n            return {\n                start: start,\n                end: end,\n                seconds: seconds,\n                displayDuration: duration,\n                displaySize: size,\n                url: '/' + vnode.attrs.channel + '/' + Math.floor(start.getTime()/1000) + '-' + Math.floor(end.getTime()/1000) + '.mp3?filename='+filename,\n            }\n        }, [vnode.state.index, vnode.state.startdate, vnode.state.starttime, vnode.state.endtime])\n        vnode.state.audioNode = m.stream(null)\n        vnode.state.playerOffset = m.stream(null)\n        vnode.state.playerResume = m.stream(false)\n        vnode.state.ontimeupdate = function() {\n            if (this !== vnode.state.audioNode())\n                return\n            var pos = this.currentTime\n            if (pos === undefined)\n                pos = null\n            if (pos === vnode.state.playerOffset())\n                return\n            if (pos !== null && Math.floor(pos) == Math.floor(vnode.state.playerOffset()))\n                return\n            vnode.state.playerOffset(pos)\n            m.redraw()\n        }\n        vnode.state.playerTime = m.stream.combine(function(want, offset) {\n            if (!want().start)\n                return null\n            return new Date(want().start.getTime() + 1000*offset())\n        }, [vnode.state.want, vnode.state.playerOffset])\n        vnode.state.want.map(function() {\n            m.route.set('/archive/:channel/:startdate/:starttime/:endtime', {\n                channel: vnode.attrs.channel,\n                startdate: vnode.state.startdate(),\n                starttime: vnode.state.starttime(),\n                endtime: vnode.state.endtime(),\n            }, {\n                replace: true,\n            })\n        })\n        vnode.state.iframe = m.stream({})\n    },\n    view: function(vnode) {\n        return [\n            m('.mdc-layout-grid', [\n                m('.mdc-layout-grid__inner', [\n                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-4', [\n                        m('div', {style: {width: '272px'}}, [\n                            m(DatePicker, {\n                                date: vnode.state.startdate,\n                                daysAvailable: vnode.state.daysAvailable(),\n                            }),\n                        ]),\n                    ]),\n                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-8', [\n                        m('.mdc-layout-grid__inner', [\n                            m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-6', [\n                                m(TextField, {\n                                    id: 'starttime',\n                                    label: 'start time',\n                                    icon: 'timer',\n                                    store: vnode.state.starttime,\n                                    style: {marginTop: 0},\n                                }),\n                                m(useCurrent, {\n                                    disabled: vnode.state.playerOffset()===null || !vnode.state.want().url,\n                                    dst: function(t) {\n                                        vnode.state.starttime(t)\n                                        vnode.state.playerResume(true)\n                                    },\n                                    src: vnode.state.playerTime,\n                                })\n                            ]),\n                            m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-6', [\n                                m(TextField, {\n                                    id: 'endtime',\n                                    label: 'end time',\n                                    icon: 'timer',\n                                    store: vnode.state.endtime,\n                                    style: {marginTop: 0},\n                                }),\n                                m(useCurrent, {\n                                    disabled: vnode.state.playerOffset()===null || !vnode.state.want().url,\n                                    dst: vnode.state.endtime,\n                                    src: vnode.state.playerTime,\n                                }),\n                            ]),\n                            m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-2', {style: {textAlign: 'right'}}, [\n                                m(Clockface, {\n                                    width: 100,\n                                    height: 100,\n                                    time: vnode.state.playerTime(),\n                                }),\n                            ]),\n                            m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-10', [\n                                m('.', {style: {marginBottom: '1em'}}, [\n                                    m('audio', {\n                                        oncreate: function(audioNode) {\n                                            vnode.state.playerOffset(null)\n                                            vnode.state.audioNode(audioNode.dom)\n                                            audioNode.state.onkeydown = function(e) {\n                                                if (document.activeElement.tagName == 'INPUT')\n                                                    return\n                                                else if (e.key === 'ArrowRight')\n                                                    audioNode.dom.currentTime = Math.max(Math.floor((audioNode.dom.currentTime+5)/5)*5, 0)\n                                                else if (e.key === 'ArrowLeft')\n                                                    audioNode.dom.currentTime = Math.max(Math.floor((audioNode.dom.currentTime-1)/5)*5, 0)\n                                                else if (e.key === ' ')\n                                                    audioNode.dom.paused ? audioNode.dom.play() : audioNode.dom.pause()\n                                                else\n                                                    return\n                                                e.preventDefault()\n                                            }\n                                            document.body.addEventListener('keydown', audioNode.state.onkeydown, {capture: true})\n                                        },\n                                        onremove: function(audioNode) {\n                                            document.body.removeEventListener('keydown', audioNode.state.onkeydown, {capture: true})\n                                        },\n                                        ondurationchange: vnode.state.ontimeupdate,\n                                        onemptied: vnode.state.ontimeupdate,\n                                        onended: vnode.state.ontimeupdate,\n                                        onabort: vnode.state.ontimeupdate,\n                                        ontimeupdate: vnode.state.ontimeupdate,\n                                        controls: true,\n                                        controlsList: 'nodownload',\n                                        preload: 'metadata',\n                                        style: {\n                                            width: '100%',\n                                        },\n                                    }, [\n                                        vnode.state.want().url && m('source', {\n                                            onupdate: function(vnode) {\n                                                var audio = vnode.dom.parentElement\n                                                if (vnode.state.url !== vnode.attrs.channel) {\n                                                    vnode.state.url = vnode.attrs.channel\n                                                    audio.autoplay = vnode.attrs.resume() && audio.buffered.length>0 && !audio.paused\n                                                    audio.load()\n                                                    vnode.attrs.resume(false)\n                                                }\n                                            },\n                                            onbeforeremove: function(vnode) {\n                                                vnode.state.parent = vnode.dom.parentElement\n                                            },\n                                            onremove: function(vnode) {\n                                                if (vnode.state.parent)\n                                                    vnode.state.parent.load()\n                                            },\n                                            key: vnode.state.want().url,\n                                            src: vnode.state.want().url,\n                                            resume: vnode.state.playerResume,\n                                        }),\n                                    ]),\n                                ]), m('.', [\n                                    m(Button, {\n                                        disabled: !vnode.state.want().url,\n                                        raised: true,\n                                        label: 'download',\n                                        icon: 'file_download',\n                                        onclick: function() {\n                                            vnode.state.iframe().src = vnode.state.want().url\n                                        },\n                                    }),\n                                    !vnode.state.want().seconds ? null : m('span', {style: {marginLeft: '2em'}}, [\n                                        vnode.state.want().displayDuration,\n                                        m.trust(' &mdash; '),\n                                        vnode.state.want().displaySize,\n                                    ]),\n                                ]),\n                            ]),\n                        ]),\n                    ]),\n                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-8', {\n                        oncreate: function(cell) {\n                            vnode.state.mapWidth = cell.dom.getBoundingClientRect().width\n                            m.redraw()\n                        },\n                        onupdate: function(cell) {\n                            var cw = cell.dom.getBoundingClientRect().width\n                            if (vnode.state.mapWidth === cw) return\n                            vnode.state.mapWidth = cw\n                            m.redraw()\n                        },\n                    }, [\n                        m(IntervalMap, {\n                            index: vnode.state.index(),\n                            width: vnode.state.mapWidth,\n                            height: 24 * IntervalMap.Days(vnode.state.index().intervals),\n                            rowHeight: 24,\n                            selection: vnode.state.want().start ? [vnode.state.want().start, vnode.state.want().end] : null,\n                        }),\n                    ]),\n                ]),\n            ]),\n            m('iframe[width=0][height=0]', {\n                oncreate: function(v) {\n                    vnode.state.iframe(v.dom)\n                },\n            }),\n        ]\n    },\n}\nvar Layout = {\n    oninit: function(vnode) {\n        vnode.state.drawer = null\n        vnode.state.channels = m.stream({})\n        m.request('/sys/channels').then(vnode.state.channels)\n        vnode.state.theme = m.stream({})\n        m.request('/sys/theme').then(vnode.state.theme)\n    },\n    view: function(vnode) {\n\treturn [\n\t    m('header.mdc-top-app-bar', {\n                oncreate: MDC.create.bind(null, mdc.topAppBar.MDCTopAppBar),\n                onremove: MDC.remove,\n            }, [\n                m('.mdc-top-app-bar__row', [\n                    m('section.mdc-top-app-bar__section.mdc-top-app-bar__section--align-start', [\n                        m('a[href=#].material-icons.mdc-top-app-bar__navigation-icon', {\n                            onclick: function() {\n                                vnode.state.drawer.open = true\n                                return false\n                            },\n                        }, 'menu'),\n\t\t        m('span.mdc-top-app-bar__title', vnode.state.theme().title || 'Audio archive'),\n                    ]),\n\t        ]),\n\t    ]),\n            m('aside.mdc-drawer.mdc-drawer--temporary.mdc-typography', {\n                oncreate: function(drawernode) {\n                    vnode.state.drawer = MDC.create(mdc.drawer.MDCTemporaryDrawer, drawernode)\n                    if (m.route.get() === '/')\n                        vnode.state.drawer.open = true\n                },\n                onremove: MDC.remove,\n            }, [\n                m('nav.mdc-drawer__drawer', [\n                    m('header.mdc-drawer__header', [\n                        m('.mdc-drawer__header-content', 'Available channels:'),\n                    ]),\n                    m('nav.mdc-drawer__content.mdc-list', [\n                        Object.keys(vnode.state.channels()).sort().map(function(name) {\n                            if (name.indexOf('/')!=0 || !vnode.state.channels()[name].archive)\n                                return null\n                            return m('a.mdc-list-item', {\n                                className: m.route.param('channel')==name.slice(1) && 'mdc-list-item--activated',\n                                href: '/archive'+name,\n                                oncreate: m.route.link,\n                                onclick: function() { vnode.state.drawer.open = false },\n                            }, [\n                                m('i.material-icons.mdc-list-item__graphic[aria-hidden=true]', 'music_note'),\n                                name,\n                            ])\n                        }),\n                    ]),\n                ]),\n            ]),\n\t    vnode.children,\n\t]\n    },\n}\nvar ArchiveRoute = {\n    render: function(vnode) {\n        return m(Layout, [\n            m(ArchivePage, Object.assign({}, vnode.attrs, {key: vnode.attrs.channel})),\n        ])\n    },\n}\nm.route(document.body, \"/\", {\n    \"/\": Layout,\n    \"/archive/:channel\": ArchiveRoute,\n    \"/archive/:channel/:startdate/:starttime/:endtime\": ArchiveRoute,\n})\nwindow.onresize = m.redraw\n"}},
	"/index.html": {N: "index.html", M: 0x1a4, S: 764, MT: 1525202415259677074, FileData: canfs.StringData{Data: "<!doctype html>\n<html lang=\"en\">\n  <head>\n    <meta charset=\"utf-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1, shrink-to-fit=no\">\n    <link rel=\"stylesheet\" href=\"https://unpkg.com/material-components-web@0.34.1/dist/material-components-web.css\">\n    <link rel=\"stylesheet\" href=\"https://fonts.googleapis.com/icon?family=Material+Icons\">\n    <title>archive</title>\n  </head>\n  <body class=\"mdc-typography\" style=\"margin: 0;\">\n    <script src=\"https://unpkg.com/material-components-web@0.34.1/dist/material-components-web.min.js\"></script>\n    <script src=\"https://unpkg.com/mithril@1.1.6/mithril.js\"></script>\n    <script src=\"https://unpkg.com/mithril-stream@1.1.0\"></script>\n    <script src=\"archive.js\"></script>\n  </body>\n</html>\n"}},
}}

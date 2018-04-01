var MP3Dir = {
    seconds: function(index, start, end) {
        var s = 0
        start = start.getTime()/1000
        end = end.getTime()/1000
        index.intervals.forEach(function(interval) {
            var istart = interval[0]
            var isec = interval[1]
            if (istart > end) return
            if (istart + isec < start) return
            if (istart + isec > end)
                isec = end - istart
            if (start > istart)
                isec -= (start - istart)
            s += isec
        })
        return s
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
            m('i.material-icons.mdc-text-field__icon[tabindex=0]', vnode.attrs.icon),
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
var ArchivePage = {
    Play: function() {
        this.audio.src = this.want().url
        this.audio.play()
    },
    Stop: function() {
        this.audio.pause()
        this.audio.removeAttribute('src')
    },
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
                seconds: seconds,
                displayDuration: duration,
                displaySize: size,
                url: vnode.state.src() + '/' + Math.floor(start.getTime()/1000) + '-' + Math.floor(end.getTime()/1000) + '.mp3?filename='+filename,
            }
        }, [vnode.state.index, vnode.state.startdate, vnode.state.starttime, vnode.state.endtime])
        vnode.state.audio = document.createElement('audio')
        Object.assign(vnode.state.audio, {
            onplaying: m.redraw,
            onpause: m.redraw,
            onprogress: m.redraw,
            onratechange: m.redraw,
            autoplay: false,
        })
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
                    ]),
                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-12', [
                        m(TextField, {
                            id: 'endtime',
                            label: 'end time',
                            icon: 'timer',
                            store: vnode.state.endtime,
                        }),
                    ]),
                    m('.mdc-layout-grid__cell.mdc-layout-grid__cell--span-12', [
                        m(Button, {
                            disabled: (!vnode.state.audio.src || !vnode.state.audio.paused) && !vnode.state.want().url,
                            label: 'preview',
                            icon: !vnode.state.audio.paused ? 'pause_circle_outline' : 'play_circle_outline',
                            onclick: function() {
                                if (this.audio.paused || this.audio.ended)
                                    this.Play()
                                else
                                    this.Stop()
                            }.bind(vnode.state),
                        }),
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
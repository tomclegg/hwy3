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
        return m('button.mdc-button', {
            className: vnode.attrs.raised && 'mdc-button--raised',
            disabled: vnode.attrs.disabled || null,
            onclick: vnode.attrs.action,
        }, [
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
    },
    oninit: function(vnode) {
        var def = new Date((new Date()).getTime() - 86400000)
        vnode.state.src = m.stream('/test')
        vnode.state.startdate = m.stream(''+def.getFullYear()+'-'+(def.getMonth()+1).toString().padStart(2,'0')+'-'+def.getDate().toString().padStart(2,'0'))
        vnode.state.starttime = m.stream(def.toLocaleTimeString())
        vnode.state.endtime = m.stream((new Date(def.getTime() + 1800000)).toLocaleTimeString())
        vnode.state.want = m.stream.combine(function(startdate, starttime, endtime) {
            try {
                if (!/^[0-9]+-[0-9]+-[0-9]+$/.test(startdate()))
                    return {error: 'no date: '+startdate()}
                var d = startdate().split('-')
                var start = new Date(d[0], d[1]-1, d[2])
                var end = new Date(start.getTime())
                var t = starttime().split(':')
                start.setHours(t[0])
                start.setMinutes(t[1])
                start.setSeconds(t[2])
                t = endtime().split(':')
                end.setHours(t[0])
                end.setMinutes(t[1])
                end.setSeconds(t[2])
                if (end < start)
                    end.setDate(end.getDate()+1)
                if (end <= start)
                    return {error: 'negative interval?'}
                return {
                    url: vnode.state.src() + '/' + Math.floor(start.getTime()/1000) + '-' + Math.floor(end.getTime()/1000) + '.mp3',
                }
            } catch(e) {
                return {error: e}
            }
        }, [vnode.state.startdate, vnode.state.starttime, vnode.state.endtime])
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
                            disabled: !vnode.state.want().url,
                            raised: true,
                            label: 'download',
                            icon: 'file_download',
                            action: function() {
                                vnode.state.iframe().src = vnode.state.want().url
                            },
                        }),
                        m(Button, {
                            disabled: !vnode.state.audio.paused && !vnode.state.want().url,
                            label: 'preview',
                            icon: !vnode.state.audio.paused ? 'pause_circle_outline' : 'play_circle_outline',
                            action: function() {
                                if (this.audio.paused || this.audio.ended)
                                    this.Play()
                                else
                                    this.Stop()
                            }.bind(vnode.state),
                        }),
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

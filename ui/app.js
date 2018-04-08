var stats = m.stream({})
function updateStats() {
    if (window.document.visibilityState == 'visible')
        m.request('/sys/stats').then(stats)
}
window.setInterval(updateStats, 2000)
updateStats()

var channels = stats.map(function(stats) {
    var scale = 1000000000/stats.SampleIntervalNS
    var now = new Date().valueOf()
    var channels = []
    for (var ch in stats.ChannelBytes) {
        var data = []
        var total = 0
        var avg = 0
        stats.ChannelBytes[ch].slice(1).forEach(function(count, i) {
            total+=count
            if (count>0) avg=i+1
        })
        if (avg>0) avg = Math.floor(total/avg*scale)
        stats.ChannelBytes[ch].slice(1).forEach(function(count, i) {
            data.unshift([new Date(now-(i+1)*stats.SampleIntervalNS/1000000), count*scale, avg])
        })
        channels.push({
            key: ch,
            data: data,
            options: {
                labels: ['time', 'inst', 'avg'],
            },
        })
    }
    return channels
})

var dash = {
    view: function(vnode) {
        return (
            m('.channels', channels().map(function(channel) {
                return m('div', {
                    key: channel.key,
                }, [
                    m('h3.label', m('a', {
                        id: channel.key,
                        href: channel.key.indexOf('output:/')==0 && channel.key.slice(7),
                        target: '_blank',
                    }, channel.key)),
                    m('.graph', {
                        channel: channel,
                        style: {
                            width: '800px',
                            height: '100px',
                        },
                        oncreate: function(vnode) {
                            vnode.state.g = new Dygraph(vnode.dom, vnode.attrs.channel.data, vnode.attrs.channel.options)
                        },
                        onupdate: function(vnode) {
                            if (window.document.visibilityState == 'visible')
                                vnode.state.g.updateOptions({file: vnode.attrs.channel.data})
                            // else https://github.com/danvk/dygraphs/issues/827
                        },
                        onremove: function(vnode) {
                            vnode.state.g.destroy()
                        },
                    }),
                ])
            })))
    },
}

m.route(document.body, "/", {
    "/": dash,
})

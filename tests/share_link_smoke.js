const assert = require('assert');
const fs = require('fs');
const path = require('path');
const vm = require('vm');

const repoRoot = path.resolve(__dirname, '..');
const context = vm.createContext({
    console,
    URL,
    Map,
    Set,
    Date,
    Math,
    Buffer,
    Base64: {
        encode(str) {
            return Buffer.from(String(str), 'utf8').toString('base64');
        },
    },
});

function loadScript(relPath) {
    const source = fs.readFileSync(path.join(repoRoot, relPath), 'utf8');
    vm.runInContext(source, context, { filename: relPath });
}

loadScript('web/assets/js/util/common.js');
loadScript('web/assets/js/util/utils.js');
loadScript('web/assets/js/model/xray.js');
vm.runInContext(`
globalThis.__shareLinkExports = {
  Inbound,
  Protocols,
};
`, context);

const { Inbound, Protocols } = context.__shareLinkExports;

function buildTrojan(network) {
    const inbound = new Inbound();
    inbound.protocol = Protocols.TROJAN;
    inbound.settings = Inbound.Settings.getSettings(Protocols.TROJAN);
    inbound.settings.clients[0].password = 'trojan-password';
    inbound.port = 443;
    inbound.stream.network = network;
    inbound.stream.security = 'tls';
    inbound.stream.tls.server = 'panel.example.com';
    return inbound;
}

{
    const inbound = buildTrojan('tcp');
    const link = inbound.genLink('1.2.3.4', 'trojan-raw');
    const url = new URL(link);
    assert.equal(url.protocol, 'trojan:');
    assert.equal(url.searchParams.get('security'), 'tls');
    assert.equal(url.searchParams.get('type'), 'tcp');
    assert.equal(url.searchParams.get('sni'), 'panel.example.com');
}

{
    const inbound = buildTrojan('ws');
    inbound.stream.ws.path = '/ws';
    inbound.stream.ws.addHeader('Host', 'cdn.example.com');
    const link = inbound.genLink('1.2.3.4', 'trojan-ws');
    const url = new URL(link);
    assert.equal(url.searchParams.get('type'), 'ws');
    assert.equal(url.searchParams.get('host'), 'cdn.example.com');
    assert.equal(url.searchParams.get('path'), '/ws');
    assert.equal(url.searchParams.get('security'), 'tls');
}

{
    const inbound = buildTrojan('grpc');
    inbound.stream.grpc.serviceName = 'grpc-service';
    const link = inbound.genLink('1.2.3.4', 'trojan-grpc');
    const url = new URL(link);
    assert.equal(url.searchParams.get('type'), 'grpc');
    assert.equal(url.searchParams.get('serviceName'), 'grpc-service');
    assert.equal(url.searchParams.get('security'), 'tls');
}

{
    const inbound = new Inbound();
    inbound.protocol = Protocols.SHADOWSOCKS;
    inbound.settings = Inbound.Settings.getSettings(Protocols.SHADOWSOCKS);
    inbound.settings.method = 'aes-256-gcm';
    inbound.settings.password = 'secret';
    inbound.port = 8388;
    inbound.stream.network = 'tcp';
    inbound.stream.security = 'none';
    const link = inbound.genLink('1.2.3.4', 'ss-plain');
    assert.ok(link.startsWith('ss://'));
    assert.equal(inbound.getShareLinkWarning(), '');
}

{
    const inbound = new Inbound();
    inbound.protocol = Protocols.SHADOWSOCKS;
    inbound.settings = Inbound.Settings.getSettings(Protocols.SHADOWSOCKS);
    inbound.settings.method = 'aes-256-gcm';
    inbound.settings.password = 'secret';
    inbound.port = 8388;
    inbound.stream.network = 'ws';
    inbound.stream.security = 'tls';
    inbound.stream.tls.server = 'panel.example.com';
    inbound.stream.ws.path = '/ss';
    assert.equal(inbound.genLink('1.2.3.4', 'ss-complex'), '');
    assert.notEqual(inbound.getShareLinkWarning(), '');
}

console.log('share_link_smoke: ok');

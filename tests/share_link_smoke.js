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
        decode(str) {
            return Buffer.from(String(str), 'base64').toString('utf8');
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

function decodeVmess(link) {
    assert.ok(link.startsWith('vmess://'));
    return JSON.parse(Buffer.from(link.slice('vmess://'.length), 'base64').toString('utf8'));
}

function decodeSS(link) {
    assert.ok(link.startsWith('ss://'));
    const withoutScheme = link.slice('ss://'.length);
    const encoded = withoutScheme.split('#')[0];
    let normalized = encoded.replace(/-/g, '+').replace(/_/g, '/');
    while (normalized.length % 4 !== 0) {
        normalized += '=';
    }
    return Buffer.from(normalized, 'base64').toString('utf8');
}

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

function buildVLESS(network = 'ws') {
    const inbound = new Inbound();
    inbound.protocol = Protocols.VLESS;
    inbound.settings = Inbound.Settings.getSettings(Protocols.VLESS);
    inbound.port = 8443;
    inbound.stream.network = network;
    inbound.stream.security = 'tls';
    inbound.stream.tls.server = 'panel.example.com';
    return inbound;
}

{
    assert.equal(Inbound.normalizeShareAddress('', false), '');
    assert.equal(Inbound.normalizeShareAddress('  https://demo.example.com/path?q=1#hash  ', false), 'demo.example.com');
    assert.equal(Inbound.normalizeShareAddress('demo.example.com:12345', false), 'demo.example.com');
    assert.equal(Inbound.getCustomShareAddressError('hk.example.com'), '');
    assert.equal(Inbound.getCustomShareAddressError('1.2.3.4'), '');
    assert.equal(Inbound.getCustomShareAddressError('2001:db8::1'), '当前自定义分享地址暂不支持 IPv6，请使用域名或 IPv4');
    assert.equal(Inbound.getCustomShareAddressError('[2001:db8::1]'), '当前自定义分享地址暂不支持 IPv6，请使用域名或 IPv4');
    assert.equal(Inbound.normalizeShareAddress('https://[2001:db8::1]:8443/path?q=1#hash', false), '');
    assert.equal(Inbound.normalizeShareAddress('https://[2001:db8::1]:8443/path?q=1#hash', true), '');
}

{
    const inbound = new Inbound();
    inbound.protocol = Protocols.VMESS;
    inbound.settings = Inbound.Settings.getSettings(Protocols.VMESS);
    inbound.settings.vmesses[0].id = '11111111-1111-1111-1111-111111111111';
    inbound.settings.vmesses[0].alterId = 0;
    inbound.port = 443;
    inbound.stream.network = 'ws';
    inbound.stream.security = 'tls';
    inbound.stream.tls.server = 'panel.example.com';
    inbound.stream.ws.path = '/vmess';
    inbound.stream.ws.addHeader('Host', 'cdn.example.com');

    const original = inbound.genLink('198.51.100.10', 'vmess-original');
    const originalPayload = decodeVmess(original);
    assert.equal(originalPayload.add, 'panel.example.com');
    assert.equal(originalPayload.host, 'cdn.example.com');
    assert.equal(originalPayload.path, '/vmess');
    assert.equal(originalPayload.tls, 'tls');

    const overridden = inbound.genLink('198.51.100.10', 'vmess-original', 'https://edge.example.com/path?q=1');
    const overriddenPayload = decodeVmess(overridden);
    assert.equal(overriddenPayload.add, 'edge.example.com');
    assert.equal(overriddenPayload.host, 'cdn.example.com');
    assert.equal(overriddenPayload.path, '/vmess');
    assert.equal(overriddenPayload.tls, 'tls');
    assert.ok(overridden.startsWith('vmess://'));

    const ipv6Override = inbound.genLink('198.51.100.10', 'vmess-original', '2001:db8::8');
    const ipv6Payload = decodeVmess(ipv6Override);
    assert.equal(ipv6Payload.add, 'panel.example.com');
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
    const inbound = buildVLESS('ws');
    inbound.stream.ws.path = '/vless';
    inbound.stream.ws.addHeader('Host', 'ws-host.example.com');
    const original = inbound.genLink('203.0.113.5', 'vless-ws');
    const originalUrl = new URL(original);
    assert.equal(originalUrl.hostname, 'panel.example.com');
    assert.equal(originalUrl.searchParams.get('sni'), 'panel.example.com');
    assert.equal(originalUrl.searchParams.get('host'), 'ws-host.example.com');
    assert.equal(originalUrl.searchParams.get('path'), '/vless');

    const overridden = inbound.genLink('203.0.113.5', 'vless-ws', '198.51.100.20/');
    const overriddenUrl = new URL(overridden);
    assert.equal(overriddenUrl.hostname, '198.51.100.20');
    assert.equal(overriddenUrl.searchParams.get('sni'), 'panel.example.com');
    assert.equal(overriddenUrl.searchParams.get('host'), 'ws-host.example.com');
    assert.equal(overriddenUrl.searchParams.get('path'), '/vless');
}

{
    const inbound = buildVLESS('tcp');
    inbound.stream.security = 'xtls';
    inbound.settings.vlesses[0].flow = 'xtls-rprx-vision';
    const original = inbound.genLink('203.0.113.5', 'vless-xtls');
    const overridden = inbound.genLink('203.0.113.5', 'vless-xtls', '2001:db8::99');
    assert.equal(overridden, original);
    const url = new URL(overridden);
    assert.equal(url.searchParams.get('flow'), 'xtls-rprx-vision');
    assert.equal(url.searchParams.get('security'), 'xtls');
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

    const same = inbound.genLink('1.2.3.4', 'ss-plain', '');
    assert.equal(same, link);

    const overridden = inbound.genLink('1.2.3.4', 'ss-plain', 'ss.example.com/path?q=1');
    assert.ok(decodeSS(overridden).includes('@ss.example.com:8388'));

    const overriddenIPv6 = inbound.genLink('1.2.3.4', 'ss-plain', '2001:db8::7');
    assert.ok(decodeSS(overriddenIPv6).includes('@1.2.3.4:8388'));
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

{
    const inbound = buildTrojan('ws');
    inbound.stream.ws.path = '/trojan';
    inbound.stream.ws.addHeader('Host', 'trojan-host.example.com');
    const original = inbound.genLink('198.51.100.1', 'trojan-custom');
    const same = inbound.genLink('198.51.100.1', 'trojan-custom', '');
    assert.equal(same, original);

    const overridden = inbound.genLink('198.51.100.1', 'trojan-custom', '2001:db8::5');
    assert.ok(overridden.includes('trojan://trojan-password@panel.example.com:443'));
    const url = new URL(overridden);
    assert.equal(url.searchParams.get('sni'), 'panel.example.com');
    assert.equal(url.searchParams.get('host'), 'trojan-host.example.com');
    assert.equal(url.searchParams.get('path'), '/trojan');
}

{
    const inbound = buildVLESS('ws');
    inbound.stream.ws.path = '/qr';
    inbound.stream.ws.addHeader('Host', 'qr-host.example.com');
    const original = inbound.genLink('203.0.113.10', 'qr-vless');
    const overridden = Inbound.overrideShareLinkAddress(original, 'https://edge-qr.example.com/test');
    const unchangedIPv6 = Inbound.overrideShareLinkAddress(original, '[2001:db8::1]');
    const originalUrl = new URL(original);
    const overriddenUrl = new URL(overridden);
    assert.equal(overriddenUrl.hostname, 'edge-qr.example.com');
    assert.equal(overriddenUrl.searchParams.get('sni'), originalUrl.searchParams.get('sni'));
    assert.equal(overriddenUrl.searchParams.get('host'), originalUrl.searchParams.get('host'));
    assert.equal(overriddenUrl.searchParams.get('path'), originalUrl.searchParams.get('path'));
    assert.equal(unchangedIPv6, original);
}

{
    const inbound = new Inbound();
    inbound.protocol = Protocols.VMESS;
    inbound.settings = Inbound.Settings.getSettings(Protocols.VMESS);
    inbound.settings.vmesses[0].id = '22222222-2222-2222-2222-222222222222';
    inbound.port = 8080;
    const original = inbound.genLink('198.51.100.3', 'qr-vmess');
    const overridden = Inbound.overrideShareLinkAddress(original, 'https://203.0.113.88/demo');
    assert.equal(decodeVmess(overridden).add, '203.0.113.88');
    assert.equal(decodeVmess(original).port, decodeVmess(overridden).port);
}

console.log('share_link_smoke: ok');

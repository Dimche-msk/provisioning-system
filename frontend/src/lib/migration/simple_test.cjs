const fs = require('fs');
const { XMLParser } = require('fast-xml-parser');

const parser = new XMLParser({
    ignoreAttributes: false,
    preserveOrder: false,
    trimValues: true,
});

function parseCisco(content) {
    const jsonObj = parser.parse(content);
    const flat = {};
    function flatten(obj, prefix = '') {
        if (typeof obj !== 'object' || obj === null) {
            if (prefix) flat[prefix.slice(0, -1)] = String(obj);
            return;
        }
        for (const key in obj) {
            flatten(obj[key], prefix + key + '.');
        }
    }
    const rootKey = Object.keys(jsonObj)[0];
    if (rootKey) flatten(jsonObj[rootKey]);
    else flatten(jsonObj);
    return flat;
}

const ciscoPath = '../sample_config/real_conf/cisco_yealink/spaccef485d919b.xml';
if (fs.existsSync(ciscoPath)) {
    const content = fs.readFileSync(ciscoPath, 'utf8');
    const flat = parseCisco(content);
    console.log('Cisco Flat Keys Sample:');
    Object.keys(flat).slice(0, 10).forEach(k => console.log(`  ${k}: ${flat[k]}`));
} else {
    console.log('Cisco sample not found');
}

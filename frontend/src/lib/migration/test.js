import fs from 'fs';
import { ConfigParser, DiscoveryEngine } from './engine.js';

const ciscoFiles = [
    '../../../../sample_config/real_conf/cisco_yealink/spaccef485d919b.xml',
    '../../../../sample_config/real_conf/cisco_yealink/spaccef485d917e.xml'
];

const yealinkFiles = [
    '../../../../sample_config/real_conf/cisco_yealink/0015657f8c41.cfg',
    '../../../../sample_config/real_conf/cisco_yealink/0015657f8c6d.cfg'
];

function testGroup(name, paths) {
    console.log(`--- Testing ${name} ---`);
    if (paths.length === 0) return;

    const engine = new DiscoveryEngine();
    
    for (const path of paths) {
        if (!fs.existsSync(path)) {
            console.error(`File not found: ${path}`);
            continue;
        }
        const content = fs.readFileSync(path, 'utf-8');
        const filename = path.split('/').pop();
        const config = ConfigParser.parse(content, filename);
        
        console.log(`Processing ${filename}...`);
        const result = engine.processFile(config);
        
        if (result.differences.length > 0) {
            console.log(`Found ${result.differences.length} differences:`);
            result.differences.slice(0, 5).forEach(d => {
                console.log(`  ${d.key}: ${d.valueA} -> ${d.valueB}`);
            });
        }
    }
}

try {
    testGroup('Cisco', ciscoFiles);
    testGroup('Yealink', yealinkFiles);
} catch (err) {
    console.error(err);
}

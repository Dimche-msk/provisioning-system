import { XMLParser } from 'fast-xml-parser';

export type FlatConfig = Record<string, string>;

export interface Diff {
    key: string;
    valueA: string;
    valueB: string;
}

export interface Conflict {
    field: string;
    contributors: { key: string; value: string }[];
}

export interface DiscoveryStepResult {
    newKeys: string[];
    differences: Diff[];
    conflicts: Conflict[];
}

export class ConfigParser {
    static parse(content: string, filename: string): FlatConfig {
        const patterns: FlatConfig = { '__filename': filename };
        const lines = content.split('\n');

        // 1. Try to find XML/Tag-like patterns: <tag attr="val">value</tag>
        // Regex matches <ANY_TAG_WITH_ATTRS>VALUE</SAME_TAG>
        // We look for leaf nodes (value doesn't contain <)
        const tagRegex = /<([^>]+)>([^<]*)<\/([^\s>]+)>/g;
        let match;
        while ((match = tagRegex.exec(content)) !== null) {
            const fullTag = match[1]; // e.g. "user_name idx=\"1\""
            const value = match[2].trim();
            const tagName = match[3]; // e.g. "user_name"
            
            // Construct the pattern
            const pattern = `<${fullTag}>{value}</${tagName}>`;
            patterns[pattern] = value;
        }

        // 2. Try to find Key-Value patterns: key = value or key : value
        // We only process lines that don't look like parts of discovered tags
        for (const line of lines) {
            const trimmed = line.trim();
            if (!trimmed || trimmed.startsWith('#') || trimmed.startsWith('<')) continue;
            
            // Support both = and : as separators
            const separator = trimmed.includes('=') ? '=' : (trimmed.includes(':') ? ':' : null);
            if (separator) {
                const parts = trimmed.split(separator);
                if (parts.length >= 2) {
                    const key = parts[0].trim();
                    const value = parts.slice(1).join(separator).trim();
                    const pattern = `${key} ${separator} {value}`;
                    patterns[pattern] = value;
                }
            }
        }

        return patterns;
    }

    static extractMac(filename: string, pattern: string): string | null {
        if (!pattern) return null;
        // Convert pattern like "SEP{MAC}.xml" to Regex
        // {MAC} matches 12 hex characters
        const escaped = pattern.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
        const regexStr = '^' + escaped.replace('\\{MAC\\}', '([0-9a-fA-F]{12})') + '$';
        const regex = new RegExp(regexStr, 'i');
        const match = filename.match(regex);
        return match ? match[1].toLowerCase() : null;
    }
}

export interface Recommendation {
    id: number;
    key: string;
    type: 'global' | 'dynamic';
    value?: string;
    coverage: number;
    isApproved: boolean;
}

export class DiscoveryEngine {
    private baseline: FlatConfig | null = null;
    private mappedFields: Record<string, string> = {}; // configKey -> systemField
    private ignoredKeys: Set<string> = new Set();
    private globalKeys: Set<string> = new Set();
    private processedCount = 0;
    private keyStatistics: Record<string, Record<string, number>> = {};
    private allConfigs: FlatConfig[] = [];
    private macPattern: string = '';
    private savedRecommendations: Recommendation[] | null = null;
    private keyAliases: Record<string, string> = {}; // oldKey -> correctedKey

    constructor(baseline?: FlatConfig) {
        if (baseline) {
            this.baseline = baseline;
        }
    }

    setBaseline(config: FlatConfig, vendorId: string, modelId: string, domain: string) {
        config['__vendor'] = vendorId;
        config['__model_id'] = modelId;
        config['__domain'] = domain;
        this.baseline = config;
        this.allConfigs = [config];
        this.processedCount = 1;

        // Track statistics for baseline too
        for (const [key, value] of Object.entries(config)) {
            if (key.startsWith('__')) continue;
            if (!this.keyStatistics[key]) this.keyStatistics[key] = {};
            this.keyStatistics[key][value] = (this.keyStatistics[key][value] || 0) + 1;
        }
    }

    getProcessedCount() {
        return this.processedCount;
    }

    private applyAliases(config: FlatConfig, rawContent?: string): FlatConfig {
        const result: FlatConfig = { ...config };
        for (const [oldKey, newKey] of Object.entries(this.keyAliases)) {
            // If it's a simple key rename in the discovered patterns
            if (result[oldKey] !== undefined) {
                result[newKey] = result[oldKey];
                delete result[oldKey];
            } 
            // If it's a pattern override (newKey contains {value}) and we have raw content
            else if (newKey.includes('{value}') && rawContent) {
                try {
                    // Escape special characters but preserve our placeholder
                    const escapedPattern = newKey.replace(/[.*+?^${}()|[\]\\]/g, (m) => m === '{' || m === '}' ? m : '\\' + m);
                    const regexStr = escapedPattern.replace('{value}', '(.*?)');
                    const regex = new RegExp(regexStr, 's');
                    const match = rawContent.match(regex);
                    if (match) {
                        result[newKey] = match[1].trim();
                    }
                } catch (e) {
                    console.error("Regex extraction failed for pattern:", newKey, e);
                }
            }
        }
        return result;
    }

    ingestConfig(config: FlatConfig, vendorId: string, modelId: string, domain: string, rawContent?: string) {
        config = this.applyAliases(config, rawContent);
        config['__vendor'] = vendorId;
        config['__model_id'] = modelId;
        config['__domain'] = domain;
        this.processedCount++;
        this.allConfigs.push(config);

        // Track statistics (ignore internal keys)
        for (const [key, value] of Object.entries(config)) {
            if (key.startsWith('__')) continue;
            if (!this.keyStatistics[key]) this.keyStatistics[key] = {};
            this.keyStatistics[key][value] = (this.keyStatistics[key][value] || 0) + 1;
        }
    }

    getAllConfigs(): FlatConfig[] {
        return this.allConfigs;
    }

    removeConfig(filename: string) {
        const index = this.allConfigs.findIndex(c => c['__filename'] === filename);
        if (index === -1) return;

        const config = this.allConfigs[index];
        this.allConfigs.splice(index, 1);
        this.processedCount--;

        // Update statistics
        for (const [key, value] of Object.entries(config)) {
            if (this.keyStatistics[key] && this.keyStatistics[key][value] !== undefined) {
                this.keyStatistics[key][value]--;
                if (this.keyStatistics[key][value] <= 0) {
                    delete this.keyStatistics[key][value];
                }
                if (Object.keys(this.keyStatistics[key]).length === 0) {
                    delete this.keyStatistics[key];
                }
            }
        }

        // If we removed the baseline (index 0 potentially, though setBaseline is distinct)
        // We don't have a clean way to "re-baseline" without a full re-process, 
        // but for now we assume removal from the middle/end.
    }

    analyzeConfig(config: FlatConfig, rawContent?: string): DiscoveryStepResult {
        config = this.applyAliases(config, rawContent);
        if (!this.baseline) {
            return { newKeys: Object.keys(config), differences: [], conflicts: [] };
        }

        const newKeys: string[] = [];
        const differences: Diff[] = [];

        // Find new keys and differences
        const allKeys = new Set([...Object.keys(this.baseline), ...Object.keys(config)]);

        for (const key of allKeys) {
            if (this.ignoredKeys.has(key) || key.startsWith('__')) continue;

            const valA = this.baseline[key];
            const valB = config[key];

            if (valA === undefined && valB !== undefined) {
                newKeys.push(key);
            } else if (valA !== undefined && valB !== undefined && valA !== valB) {
                differences.push({ key, valueA: valA, valueB: valB });
            }
        }

        return { newKeys, differences, conflicts: this.validateMapping(config) };
    }

    validateMapping(config: FlatConfig): Conflict[] {
        const fieldMap: Record<string, { key: string; value: string }[]> = {};
        const conflicts: Conflict[] = [];

        for (const [key, field] of Object.entries(this.mappedFields)) {
            if (config[key] !== undefined) {
                if (!fieldMap[field]) fieldMap[field] = [];
                fieldMap[field].push({ key, value: config[key] });
            }
        }

        for (const [field, contributors] of Object.entries(fieldMap)) {
            const values = new Set(contributors.map(c => c.value));
            if (values.size > 1) {
                conflicts.push({ field, contributors });
            }
        }

        return conflicts;
    }

    mapField(configKey: string, systemField: string) {
        this.mappedFields[configKey] = systemField;
    }

    ignoreKey(configKey: string) {
        this.ignoredKeys.add(configKey);
    }

    setMacPattern(pattern: string) {
        this.macPattern = pattern;
    }

    getMacPattern(): string {
        return this.macPattern;
    }

    getExtractedData(config: FlatConfig): Record<string, string> {
        const data: Record<string, string> = {
            'phone.raw_filename': config['__filename'] || '',
            'phone.vendor': config['__vendor'] || '',
            'phone.model_id': config['__model_id'] || '',
            'phone.domain': config['__domain'] || '',
        };
        
        // Extract MAC using pattern if set
        if (this.macPattern && config['__filename']) {
            const mac = ConfigParser.extractMac(config['__filename'], this.macPattern);
            if (mac) data['phone.mac_address'] = mac;
        }

        for (const [configKey, systemField] of Object.entries(this.mappedFields)) {
            if (config[configKey] !== undefined) {
                let value = config[configKey];
                // If mapping to MAC and we have a pattern, try to extract it from the value
                // This handles cases where the user maps the filename manually
                if (systemField === 'phone.mac_address' && this.macPattern) {
                    const extracted = ConfigParser.extractMac(value, this.macPattern);
                    if (extracted) value = extracted;
                }
                data[systemField] = value;
            }
        }
        return data;
    }

    getAllExtractedData(): Record<string, string>[] {
        return this.allConfigs.map(c => this.getExtractedData(c));
    }

    getMappedFields(): Record<string, string> {
        return { ...this.mappedFields };
    }

    isKeyMapped(key: string): boolean {
        return !!this.mappedFields[key];
    }

    isKeyIgnored(key: string): boolean {
        return this.ignoredKeys.has(key);
    }

    getAnalysisSummary(): Recommendation[] {
        if (this.savedRecommendations) return this.savedRecommendations;

        const total = this.processedCount;
        const recommendations: Recommendation[] = [];
        let idCounter = 0;

        for (const [key, stats] of Object.entries(this.keyStatistics)) {
            if (this.ignoredKeys.has(key) || this.isKeyMapped(key)) continue;

            const values = Object.entries(stats).sort((a, b) => b[1] - a[1]);
            const [mostCommonValue, count] = values[0];
            const coverage = count / total;

            if (coverage >= 0.95) {
                recommendations.push({ id: idCounter++, key, type: 'global', value: mostCommonValue, coverage, isApproved: true });
            } else if (values.length > 1) {
                recommendations.push({ id: idCounter++, key, type: 'dynamic', coverage, isApproved: true });
            }
        }

        return recommendations;
    }

    setSavedRecommendations(recs: Recommendation[]) {
        this.savedRecommendations = recs;
    }

    renameKey(oldKey: string, newKey: string) {
        if (oldKey === newKey) return;

        // 0. Store alias
        this.keyAliases[oldKey] = newKey;

        // 1. Update Statistics
        if (this.keyStatistics[oldKey]) {
            this.keyStatistics[newKey] = { ...this.keyStatistics[oldKey] };
            delete this.keyStatistics[oldKey];
        }

        // 2. Update all ingested configs
        this.allConfigs = this.allConfigs.map(config => {
            if (config[oldKey] !== undefined) {
                const newConfig = { ...config };
                newConfig[newKey] = config[oldKey];
                delete newConfig[oldKey];
                return newConfig;
            }
            return config;
        });

        // 3. Update baseline
        if (this.baseline && this.baseline[oldKey] !== undefined) {
            this.baseline[newKey] = this.baseline[oldKey];
            delete this.baseline[oldKey];
        }

        // 4. Update mapped fields
        if (this.mappedFields[oldKey]) {
            this.mappedFields[newKey] = this.mappedFields[oldKey];
            delete this.mappedFields[oldKey];
        }

        // 5. Update ignored keys
        if (this.ignoredKeys.has(oldKey)) {
            this.ignoredKeys.delete(oldKey);
            this.ignoredKeys.add(newKey);
        }

        // 6. Update saved recommendations IDs if necessary
        if (this.savedRecommendations) {
            this.savedRecommendations = this.savedRecommendations.map(r => 
                r.key === oldKey ? { ...r, key: newKey } : r
            );
        }
    }
}

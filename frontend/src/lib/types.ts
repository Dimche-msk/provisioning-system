export interface PhoneLine {
    id?: number;
    phone_id?: number;
    type: string;
    account_number: number;
    panel_number?: number | null;
    key_number?: number | null;
    additional_info: string;
}

export interface Phone {
    id?: number;
    domain: string;
    vendor: string;
    model_id: string;
    mac_address: string | null;
    phone_number: string | null;
    ip_address: string;
    description: string;
    lines: PhoneLine[];
    expansion_module_model: string;
    expansion_modules_count?: number;
    type: string;
    model_name?: string;
    vendor_name?: string;
    created_at?: string;
    updated_at?: string;
}

export interface ModelKey {
    index: number;
    type: string;
    account?: number;
    label: string;
    x: number;
    y: number;
    my_image?: string;
    settings?: Record<string, string>;
}

export interface DeviceModel {
    id: string;
    name: string;
    vendor: string;
    type: string;
    max_account_lines: number;
    own_soft_keys: number;
    own_hard_keys: number;
    supported_expansion_modules: string[];
    maximum_expansion_modules: number;
    image: string;
    keys: ModelKey[];
    key_types: KeyType[];
    other_features: string[];
}

export interface KeyType {
    id: string;
    verbose: string;
    polygon: string;
    image: string;
}

export interface Vendor {
    id: string;
    name: string;
}

export interface DomainSettings {
    name: string;
    deploy_cmd: string;
    delete_cmd: string;
    variables: Record<string, string>;
}

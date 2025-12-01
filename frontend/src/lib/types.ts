export interface PhoneLine {
    id?: number;
    phone_id?: number;
    type: string;
    number: number;
    expansion_module_number?: number;
    key_number?: number;
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
    created_at?: string;
    updated_at?: string;
}

export interface DeviceModel {
    id: string;
    name: string;
    vendor: string;
    type: string;
    max_account_lines: number;
    OwnSoftKeys: number;
    OwnHardKeys: number;
    supported_expansion_modules: string[];
    maximum_expansion_modules: number;
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

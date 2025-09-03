export namespace db {
	
	export class Database {
	
	
	    static createFrom(source: any = {}) {
	        return new Database(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class QueryResult {
	    ok: boolean;
	    data: any;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new QueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.data = source["data"];
	        this.message = source["message"];
	    }
	}

}

export namespace main {
	
	export class AreaConfig {
	    province_name: string;
	    city_name: string;
	    country_name: string;
	
	    static createFrom(source: any = {}) {
	        return new AreaConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.province_name = source["province_name"];
	        this.city_name = source["city_name"];
	        this.country_name = source["country_name"];
	    }
	}
	export class Condition {
	    credit_code?: string;
	    stat_date?: string;
	    project_code?: string;
	    document_number?: string;
	    province_name?: string;
	    city_name?: string;
	    country_name?: string;
	
	    static createFrom(source: any = {}) {
	        return new Condition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.credit_code = source["credit_code"];
	        this.stat_date = source["stat_date"];
	        this.project_code = source["project_code"];
	        this.document_number = source["document_number"];
	        this.province_name = source["province_name"];
	        this.city_name = source["city_name"];
	        this.country_name = source["country_name"];
	    }
	}
	export class ConflictData {
	    filePath: string;
	    tableType: string;
	    conditions: Condition[];
	
	    static createFrom(source: any = {}) {
	        return new ConflictData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filePath = source["filePath"];
	        this.tableType = source["tableType"];
	        this.conditions = this.convertValues(source["conditions"], Condition);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EnvResult {
	    appName: string;
	    appFileName: string;
	    basePath: string;
	    os: string;
	    arch: string;
	    x64Level: number;
	    exePath: string;
	    assetsDir: string;
	
	    static createFrom(source: any = {}) {
	        return new EnvResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appName = source["appName"];
	        this.appFileName = source["appFileName"];
	        this.basePath = source["basePath"];
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.x64Level = source["x64Level"];
	        this.exePath = source["exePath"];
	        this.assetsDir = source["assetsDir"];
	    }
	}
	export class FileFilter {
	    name: string;
	    pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new FileFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.pattern = source["pattern"];
	    }
	}
	export class FileDialogOptions {
	    title?: string;
	    filters?: FileFilter[];
	    openDirectory?: boolean;
	    createDirectory?: boolean;
	    defaultPath?: string;
	    defaultFilename?: string;
	    multiSelections?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileDialogOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.filters = this.convertValues(source["filters"], FileFilter);
	        this.openDirectory = source["openDirectory"];
	        this.createDirectory = source["createDirectory"];
	        this.defaultPath = source["defaultPath"];
	        this.defaultFilename = source["defaultFilename"];
	        this.multiSelections = source["multiSelections"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileDialogResult {
	    canceled: boolean;
	    filePaths: string[];
	
	    static createFrom(source: any = {}) {
	        return new FileDialogResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.canceled = source["canceled"];
	        this.filePaths = source["filePaths"];
	    }
	}
	
	export class FileInfo {
	    name: string;
	    fullPath: string;
	    size: number;
	    isDirectory: boolean;
	    isFile: boolean;
	    lastModified: number;
	    ext: string;
	    parentDir: string;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.fullPath = source["fullPath"];
	        this.size = source["size"];
	        this.isDirectory = source["isDirectory"];
	        this.isFile = source["isFile"];
	        this.lastModified = source["lastModified"];
	        this.ext = source["ext"];
	        this.parentDir = source["parentDir"];
	    }
	}
	export class FlagResult {
	    ok: boolean;
	    data: string;
	
	    static createFrom(source: any = {}) {
	        return new FlagResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.data = source["data"];
	    }
	}
	export class MessageBoxOptions {
	    title?: string;
	    message: string;
	    detail?: string;
	    type?: string;
	    buttons?: string[];
	    defaultId?: number;
	    cancelId?: number;
	
	    static createFrom(source: any = {}) {
	        return new MessageBoxOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.message = source["message"];
	        this.detail = source["detail"];
	        this.type = source["type"];
	        this.buttons = source["buttons"];
	        this.defaultId = source["defaultId"];
	        this.cancelId = source["cancelId"];
	    }
	}
	export class MessageBoxResult {
	    response: number;
	    checkboxChecked: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MessageBoxResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.response = source["response"];
	        this.checkboxChecked = source["checkboxChecked"];
	    }
	}

}


export namespace data_import {
	
	export class DataImportRecord {
	    FileName: string;
	    FileType: string;
	    // Go type: time
	    ImportTime: any;
	    ImportState: string;
	    Describe: string;
	    CreateUser: string;
	
	    static createFrom(source: any = {}) {
	        return new DataImportRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.FileName = source["FileName"];
	        this.FileType = source["FileType"];
	        this.ImportTime = this.convertValues(source["ImportTime"], null);
	        this.ImportState = source["ImportState"];
	        this.Describe = source["Describe"];
	        this.CreateUser = source["CreateUser"];
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

}

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
	export class DataCheckResult {
	    ok: boolean;
	    message: string;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new DataCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.message = source["message"];
	        this.errors = source["errors"];
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
	export class ExportResult {
	    ok: boolean;
	    message: string;
	    filePath: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.message = source["message"];
	        this.filePath = source["filePath"];
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
	export class ImportProcess {
	    id: string;
	    fileName: string;
	    fileType: string;
	    status: string;
	    progress: number;
	    totalRows: number;
	    processedRows: number;
	    // Go type: time
	    startTime: any;
	    // Go type: time
	    endTime: any;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportProcess(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fileName = source["fileName"];
	        this.fileType = source["fileType"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.totalRows = source["totalRows"];
	        this.processedRows = source["processedRows"];
	        this.startTime = this.convertValues(source["startTime"], null);
	        this.endTime = this.convertValues(source["endTime"], null);
	        this.message = source["message"];
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
	export class ManualCheckResult {
	    obj_id: string;
	    table_name: string;
	    check_result: string;
	    check_remark: string;
	    check_user: string;
	    check_time: string;
	
	    static createFrom(source: any = {}) {
	        return new ManualCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.obj_id = source["obj_id"];
	        this.table_name = source["table_name"];
	        this.check_result = source["check_result"];
	        this.check_remark = source["check_remark"];
	        this.check_user = source["check_user"];
	        this.check_time = source["check_time"];
	    }
	}
	export class MergeResult {
	    ok: boolean;
	    message: string;
	    successCount: number;
	    conflictCount: number;
	    errorCount: number;
	
	    static createFrom(source: any = {}) {
	        return new MergeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.message = source["message"];
	        this.successCount = source["successCount"];
	        this.conflictCount = source["conflictCount"];
	        this.errorCount = source["errorCount"];
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


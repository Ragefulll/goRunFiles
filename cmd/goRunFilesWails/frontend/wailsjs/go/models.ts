export namespace app {
	
	export class DisplayStatus {
	    name: string;
	    type: string;
	    status: string;
	    disabled: boolean;
	    icon: string;
	    pid: string;
	    started_at: string;
	    uptime: string;
	    target: string;
	    error: string;
	    hung: boolean;
	    cpu: string;
	    gpu: string;
	    gpu_mem_mb: string;
	    mem_mb: string;
	    net_kbs: string;
	    io_kbs: string;
	
	    static createFrom(source: any = {}) {
	        return new DisplayStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.status = source["status"];
	        this.disabled = source["disabled"];
	        this.icon = source["icon"];
	        this.pid = source["pid"];
	        this.started_at = source["started_at"];
	        this.uptime = source["uptime"];
	        this.target = source["target"];
	        this.error = source["error"];
	        this.hung = source["hung"];
	        this.cpu = source["cpu"];
	        this.gpu = source["gpu"];
	        this.gpu_mem_mb = source["gpu_mem_mb"];
	        this.mem_mb = source["mem_mb"];
	        this.net_kbs = source["net_kbs"];
	        this.io_kbs = source["io_kbs"];
	    }
	}
	export class DisplaySnapshot {
	    updated: string;
	    version: string;
	    check_timing_ms: number;
	    check_process_running: boolean;
	    net_unit: string;
	    net_mode: string;
	    net_err: string;
	    net_dbg: string;
	    items: DisplayStatus[];
	
	    static createFrom(source: any = {}) {
	        return new DisplaySnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.updated = source["updated"];
	        this.version = source["version"];
	        this.check_timing_ms = source["check_timing_ms"];
	        this.check_process_running = source["check_process_running"];
	        this.net_unit = source["net_unit"];
	        this.net_mode = source["net_mode"];
	        this.net_err = source["net_err"];
	        this.net_dbg = source["net_dbg"];
	        this.items = this.convertValues(source["items"], DisplayStatus);
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

export namespace config {
	
	export class SettingsDTO {
	    checkTiming: string;
	    restartTiming: string;
	    autoRestart: boolean;
	    autoRestartTime: string;
	    autoRestartOnExit: boolean;
	    launchInNewConsole: boolean;
	    autoCloseErrorDialogs: boolean;
	    errorWindowTitles: string;
	    useETWNetwork: boolean;
	    netDebug: boolean;
	    netUnit: string;
	    netScale: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.checkTiming = source["checkTiming"];
	        this.restartTiming = source["restartTiming"];
	        this.autoRestart = source["autoRestart"];
	        this.autoRestartTime = source["autoRestartTime"];
	        this.autoRestartOnExit = source["autoRestartOnExit"];
	        this.launchInNewConsole = source["launchInNewConsole"];
	        this.autoCloseErrorDialogs = source["autoCloseErrorDialogs"];
	        this.errorWindowTitles = source["errorWindowTitles"];
	        this.useETWNetwork = source["useETWNetwork"];
	        this.netDebug = source["netDebug"];
	        this.netUnit = source["netUnit"];
	        this.netScale = source["netScale"];
	    }
	}
	export class ProcessDTO {
	    name: string;
	    disabled: boolean;
	    type: string;
	    process: string;
	    path: string;
	    command: string;
	    args: string;
	    screen: number;
	    checkProcess: string;
	    checkCmdline: string;
	    checkCmdlineExclude: string;
	    delayStartTime: string;
	    monitorHang: boolean;
	    hangTimeout: string;
	
	    static createFrom(source: any = {}) {
	        return new ProcessDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.disabled = source["disabled"];
	        this.type = source["type"];
	        this.process = source["process"];
	        this.path = source["path"];
	        this.command = source["command"];
	        this.args = source["args"];
	        this.screen = source["screen"];
	        this.checkProcess = source["checkProcess"];
	        this.checkCmdline = source["checkCmdline"];
	        this.checkCmdlineExclude = source["checkCmdlineExclude"];
	        this.delayStartTime = source["delayStartTime"];
	        this.monitorHang = source["monitorHang"];
	        this.hangTimeout = source["hangTimeout"];
	    }
	}
	export class ConfigDTO {
	    processes: ProcessDTO[];
	    settings: SettingsDTO;
	
	    static createFrom(source: any = {}) {
	        return new ConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.processes = this.convertValues(source["processes"], ProcessDTO);
	        this.settings = this.convertValues(source["settings"], SettingsDTO);
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

export namespace display {
	
	export class Screen {
	    index: number;
	    name: string;
	    primary: boolean;
	    x: number;
	    y: number;
	    width: number;
	    height: number;
	
	    static createFrom(source: any = {}) {
	        return new Screen(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.name = source["name"];
	        this.primary = source["primary"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.width = source["width"];
	        this.height = source["height"];
	    }
	}

}

export namespace main {
	
	export class TaskStatus {
	    installed: boolean;
	    running: boolean;
	    state: string;
	    lastRunTime: string;
	    lastTaskResult: number;
	    taskName: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installed = source["installed"];
	        this.running = source["running"];
	        this.state = source["state"];
	        this.lastRunTime = source["lastRunTime"];
	        this.lastTaskResult = source["lastTaskResult"];
	        this.taskName = source["taskName"];
this.error = source["error"];
    }
}

export class UpdateStatus {
    current: string;
    remote: string;
    status: string;
    progress: number;
    detail: string;

    static createFrom(source: any = {}) {
        return new UpdateStatus(source);
    }

    constructor(source: any = {}) {
        if ('string' === typeof source) source = JSON.parse(source);
        this.current = source["current"];
        this.remote = source["remote"];
        this.status = source["status"];
        this.progress = source["progress"];
        this.detail = source["detail"];
    }
}

}


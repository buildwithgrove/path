"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.timeIt = void 0;
const picocolors_1 = __importDefault(require("picocolors"));
const pretty_ms_1 = __importDefault(require("pretty-ms"));
async function timeIt(name, action) {
    const startTime = new Date();
    await action();
    console.log(`${picocolors_1.default.green(`Task ${name} done (took `)} ${picocolors_1.default.white((0, pretty_ms_1.default)(new Date() - startTime))}${picocolors_1.default.green(`)`)}`);
}
exports.timeIt = timeIt;
//# sourceMappingURL=utils.js.map
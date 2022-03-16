<!--
 * @Author: your name
 * @Date: 2022-03-07 11:21:20
 * @LastEditTime: 2022-03-10 13:44:29
 * @LastEditors: your name
 * @Description: 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 * @FilePath: /flv/README.md
-->
// Tinygo补丁，js.finalizeRef修复GC问题，需要生成wasm_exec.js后手动修改
// before
"syscall/js.finalizeRef": (sp) => {
  // Note: TinyGo does not support finalizers so this should never be
  // called.
  console.error('syscall/js.finalizeRef not implemented');
},

// after
"syscall/js.finalizeRef": (sp) => {
  // Note: TinyGo does not support finalizers so this should never be
  // called.
  // console.error('syscall/js.finalizeRef not implemented');
  // 补丁，修复GC问题
  const id = mem().getUint32(sp + 8, true);
  this._goRefCounts[id]--;
  if (this._goRefCounts[id] === 0) {
      const v = this._values[id];
      this._values[id] = null;
      this._ids.delete(v);
      this._idPool.push(id);
  }
},
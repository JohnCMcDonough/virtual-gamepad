
declare module 'mqtt-emitter' {
    
  import EventEmitter from "events";
  export default class MQTTEmitter extends EventEmitter {
      public onadd: any;
      public onremove: any;
  }
}
#import <Cocoa/Cocoa.h>
#import <Carbon/Carbon.h>
static NSArray *sources(void) {
 NSDictionary *query=@{(__bridge id)kTISPropertyBundleID:@"org.shadowinput.inputmethod.shadowinput"};
 return CFBridgingRelease(TISCreateInputSourceList((__bridge CFDictionaryRef)query,true));
}
static int registerNative(const char *path) {
 @autoreleasepool {
  NSURL *url=[NSURL fileURLWithPath:@(path)];
  OSStatus status=LSRegisterURL((__bridge CFURLRef)url,true);
  if(status==noErr)status=TISRegisterInputSource((__bridge CFURLRef)url);
  if(status!=noErr)return status;
  // Enable the parent before its mode. Selection remains the user's choice.
  NSArray *list=sources();
  if(!list.count)return paramErr;
  for(id object in list) {
   TISInputSourceRef src=(__bridge TISInputSourceRef)object;
   CFTypeRef identity=TISGetInputSourceProperty(src,kTISPropertyInputSourceID);
   if(identity && CFEqual(identity,CFSTR("org.shadowinput.inputmethod.shadowinput"))) {
    status=TISEnableInputSource(src);
    if(status!=noErr)return status;
   }
  }
  for(id object in list) {
   status=TISEnableInputSource((__bridge TISInputSourceRef)object);
   if(status!=noErr)return status;
  }
  return 0;
 }
}

int shadow_ime_register(const char *path) {
 if(NSThread.isMainThread)return registerNative(path);
 __block int result=0;
 dispatch_sync(dispatch_get_main_queue(),^{result=registerNative(path);});
 return result;
}

#import <Cocoa/Cocoa.h>
#import "typing_fixture.h"
#import "../learning.h"
int main(int argc,const char *argv[]){
 @autoreleasepool {
  if(argc!=3)return 2;
  [NSApplication sharedApplication];
  NSMutableData *keyBytes=[NSMutableData dataWithLength:32];memset(keyBytes.mutableBytes,0x61,32);
  NSError *error=nil;
  require(ArionRegisterTestEncryptedLearning([@(argv[2]) stringByAppendingPathComponent:@"learning"],keyBytes,&error),@"isolated encrypted learning");
  ArionKeyboardEngine *engine=[ArionKeyboardEngine new];
  require([engine startWithData:@(argv[1]) userDirectory:@(argv[2]) error:&error],@"real engine starts");
  ArionTyping *typing=[[ArionTyping alloc] initWithEngine:engine];TypingEditor *editor=[TypingEditor new];
  require([typing activate:editor error:&error],@"activate real editor");if(typing.ascii)[typing toggleASCII];
  type(typing,@"mingtianwomenyiqitaolun");
  require(typing.aiSource.length>=4,@"full-span Chinese candidate available");
  NSUInteger old=typing.revision;
  type(typing,@"a");
  require(![typing commitAI:@"迟到结果" revision:old] && editor.insertions==0,@"typing revokes old AI result");
  [typing cancel];type(typing,@"mingtianwomenyiqitaolun");
  require([typing commitAI:@"明天我们一起讨论" revision:typing.revision],@"explicit AI candidate commits");
  require([editor.view.string isEqual:@"明天我们一起讨论"] && editor.insertions==1 && !typing.composing,@"single commit replaces marked composition");
  require(![typing commitAI:@"重复" revision:typing.revision],@"no duplicate commit");
  type(typing,@"nihao");press(typing,@" ",49);
  require([editor.view.string hasSuffix:@"你好"] && editor.insertions==2,@"ordinary Pinyin survives AI commit");
  type(typing,@"mingtian");old=typing.revision;[typing deactivate];
  require(![typing commitAI:@"过期" revision:old],@"focus/source deactivation revokes candidate");
  [engine stop];puts("PASS: native Pinyin, full-span AI source, explicit single commit, stale revision and focus cancellation");
 }
 return 0;
}

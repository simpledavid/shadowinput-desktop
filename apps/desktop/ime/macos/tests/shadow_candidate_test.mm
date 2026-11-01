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
  require(typing.typoCorrection,@"shadowinput defaults to the engine's typo-correction mode");
  require([typing activate:editor error:&error],@"activate real editor");if(typing.ascii)[typing toggleASCII];
  type(typing,@"mingtian");
  require(!typing.aiSource.length,@"unfinished Pinyin never starts a model request");
  press(typing,@" ",49);
  require([typing.aiSource isEqual:@"明天"],@"ordinary Space commit becomes continuation context");
  NSUInteger old=typing.revision;
  type(typing,@"women");
  require(![typing commitAI:@"迟到结果" revision:old] && editor.insertions==1,@"typing revokes old AI result");
  press(typing,@" ",49);
  require([typing.aiSource isEqual:@"明天我们"],@"consecutive word commits preserve short context");
  require([typing commitAI:@"一起讨论" revision:typing.revision],@"explicit AI candidate appends");
  require([editor.view.string isEqual:@"明天我们一起讨论"] && editor.insertions==3 && !typing.composing,@"continuation preserves committed text");
  require(![typing commitAI:@"重复" revision:typing.revision],@"no duplicate commit");
  type(typing,@"nihao");press(typing,@" ",49);
  require([editor.view.string hasSuffix:@"你好"] && editor.insertions==4,@"ordinary Pinyin survives AI commit");
  old=typing.revision;[editor.view setSelectedRange:NSMakeRange(0,0)];
  require(![typing commitAI:@"错误位置" revision:old],@"moving caret without a key cancels insertion");
  require(!typing.aiSource.length,@"moved caret cannot reuse old context");
  type(typing,@"mingtian");press(typing,@" ",49);old=typing.revision;[typing deactivate];
  require(![typing commitAI:@"过期" revision:old],@"focus/source deactivation revokes candidate");
  require([typing activate:editor error:&error],@"reactivate editor");if(typing.ascii)[typing toggleASCII];
  type(typing,@"zhogn");
  require([typing.snapshot.candidates containsObject:@"中"],@"default typo correction recovers transposed Pinyin");
  [typing cancel];type(typing,@"zhong");
  require([typing.snapshot.candidates containsObject:@"中"],@"exact Pinyin remains available");
  [typing deactivate];
  [engine stop];puts("PASS: word commits, bounded continuation, append once, stale revision, caret and focus cancellation");
 }
 return 0;
}

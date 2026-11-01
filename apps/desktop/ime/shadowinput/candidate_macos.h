// Included only by the shadowinput IMK build. All UI state belongs to the main thread.
#include "continuation.h"
static BOOL shadowAIEnabled(void) {
 NSString *base=NSSearchPathForDirectoriesInDomains(NSApplicationSupportDirectory,NSUserDomainMask,YES).firstObject;
 return [NSFileManager.defaultManager fileExistsAtPath:[base stringByAppendingPathComponent:@"shadowinput/candidate.endpoint"]];
}
@interface ShadowCandidate : NSObject
@property(strong) ArionCandidatePanel *panel;
@property(strong) NSURLSession *session;
@property(strong) NSTimer *timer;
@property(strong) NSURLSessionDataTask *task;
@property NSUInteger generation;
@property NSUInteger revision;
@property(copy) NSString *text;
@property(copy) void (^ready)(void);
@property(copy) BOOL (^valid)(void);
- (void)update:(NSString *)source revision:(NSUInteger)revision;
- (void)clear;
@end
@implementation ShadowCandidate
- (instancetype)init {
 if ((self=[super init])) {
  self.panel=[ArionCandidatePanel new];
  __weak ShadowCandidate *weak=self;
  self.timer=[NSTimer scheduledTimerWithTimeInterval:0.25 repeats:YES block:^(NSTimer *timer){(void)timer;ShadowCandidate *owner=weak;if(owner && (!shadowAIEnabled() || (owner.text.length && owner.valid && !owner.valid())))[owner clear];}];
  NSURLSessionConfiguration *config=NSURLSessionConfiguration.ephemeralSessionConfiguration;
  config.URLCache=nil; config.HTTPCookieStorage=nil; config.connectionProxyDictionary=@{};
  self.session=[NSURLSession sessionWithConfiguration:config];
 }
 return self;
}
- (void)clear { self.generation++; [self.task cancel]; self.task=nil; self.text=nil; [self.panel hide]; }
- (void)dealloc { [self.timer invalidate]; [self.task cancel]; [self.session invalidateAndCancel]; }
- (void)update:(NSString *)source revision:(NSUInteger)revision {
 [self clear]; self.revision=revision;
 if (!source || shadow_runes(source.UTF8String)<2 || shadow_runes(source.UTF8String)>ShadowContextRunes) return;
 NSUInteger generation=self.generation;
 __weak ShadowCandidate *weak=self;
 dispatch_after(dispatch_time(DISPATCH_TIME_NOW,ShadowDebounceMS*NSEC_PER_MSEC),dispatch_get_main_queue(),^{
  ShadowCandidate *owner=weak; if (!owner || owner.generation!=generation) return;
  NSString *base=NSSearchPathForDirectoriesInDomains(NSApplicationSupportDirectory,NSUserDomainMask,YES).firstObject;
  NSString *file=[base stringByAppendingPathComponent:@"shadowinput/candidate.endpoint"];
  NSString *config=[NSString stringWithContentsOfFile:file encoding:NSUTF8StringEncoding error:nil];
  NSArray *lines=[config componentsSeparatedByString:@"\n"];
  if (lines.count!=3 || ![lines[0] hasPrefix:@"127.0.0.1:"] || [lines[1] length]<20) return;
  NSString *port=[lines[0] substringFromIndex:10];
  if ([port rangeOfCharacterFromSet:NSCharacterSet.decimalDigitCharacterSet.invertedSet].location!=NSNotFound || port.intValue<1 || port.intValue>65535) return;
  NSURL *url=[NSURL URLWithString:[NSString stringWithFormat:@"http://127.0.0.1:%d/candidate",port.intValue]];
  NSMutableURLRequest *request=[NSMutableURLRequest requestWithURL:url cachePolicy:NSURLRequestReloadIgnoringLocalCacheData timeoutInterval:16];
  request.HTTPMethod=@"POST"; request.HTTPBody=[source dataUsingEncoding:NSUTF8StringEncoding];
  [request setValue:[@"Bearer " stringByAppendingString:lines[1]] forHTTPHeaderField:@"Authorization"];
  [request setValue:@"text/plain; charset=utf-8" forHTTPHeaderField:@"Content-Type"];
  owner.task=[owner.session dataTaskWithRequest:request completionHandler:^(NSData *data,NSURLResponse *response,NSError *error) {
   if (error || ![response isKindOfClass:NSHTTPURLResponse.class] || ((NSHTTPURLResponse *)response).statusCode!=200 || data.length>1024) return;
   NSString *text=[[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
   if (!text.length || shadow_runes(text.UTF8String)>ShadowSuffixRunes || [text rangeOfCharacterFromSet:NSCharacterSet.controlCharacterSet].location!=NSNotFound) return;
   dispatch_async(dispatch_get_main_queue(),^{
    ShadowCandidate *live=weak; if (!live || live.generation!=generation) return;
    live.text=text; live.task=nil; if (live.ready) live.ready();
   });
  }];
  [owner.task resume];
 });
}
@end

# mapreduce

**Interface:**

- Map 对slice各元素进行映射执行function，返回新slice
- MapInplace 对slice各元素进行映射执行function，就地修改slice
- Reduce 对slice各元素进行pairFunc规约
- Filter 对slice各元素执行function过滤，返回新slice
- FilterInplace 对slice各元素执行function过滤，就地修改slice

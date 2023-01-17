#### 未解决问题
1、Merkle对Transaction序列化的方式
eos\libraries\chain\controller.cpp  2068行
```sh
trx_digests.emplace_back( a.digest() );
```



#### 已解决问题
1、Merkle树叶子节点数据类型为transactions列表中元素 transaction_receipt
逻辑为
eos\libraries\chain\controller.cpp  1848行，2065行
```sh
auto trx_mroot = calculate_trx_merkle( b->transactions );


static checksum256_type calculate_trx_merkle( const deque<transaction_receipt>& trxs ) 
```
2、Merkle树构建逻辑
eos\libraries\chain\merkle.cpp  35行~50行
```sh
digest_type merkle(deque<digest_type> ids) {
   if( 0 == ids.size() ) { return digest_type(); }

   while( ids.size() > 1 ) {
      if( ids.size() % 2 )
         ids.push_back(ids.back());

      for (size_t i = 0; i < ids.size() / 2; i++) {
         ids[i] = digest_type::hash(make_canonical_pair(ids[2 * i], ids[(2 * i) + 1]));
      }

      ids.resize(ids.size() / 2);
   }

   return ids.front();
}
```
3、构成默克尔树时，左右节点处理逻辑
eos\libraries\chain\merkle.cpp  14行~24行

```sh
digest_type make_canonical_left(const digest_type& val) {
   digest_type canonical_l = val;
   canonical_l._hash[0] &= 0xFFFFFFFFFFFFFF7FULL;
   return canonical_l;
}

digest_type make_canonical_right(const digest_type& val) {
   digest_type canonical_r = val;
   canonical_r._hash[0] |= 0x0000000000000080ULL;
   return canonical_r;
}
```

package governance

// snapshot 단위 commit, rollback 테스트
// 로딩 순서 : memory -> state db -> trie ( badger db ) -> default value 순

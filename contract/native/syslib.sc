/**
 * @file    syslib.sc
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef sc_def
#define sc_def(src)
#endif

sc_def(
library system {
    func abs32(int32 v) int32 as __abs32;
    func abs64(int64 v) int64 as __abs64;
    func abs256(int256 v) int256 as __mpz_abs;

    func pow32(int32 x, int32 y) int32 as __pow32;
    func pow64(int64 x, int32 y) int64 as __pow64;
    func pow256(int256 x, int32 y) int256 as __mpz_pow;

    /*
    func sign32(int32 v) int8 { return v > 0 ? 1 : (v < 0 ? -1 : 0); }
    func sign64(int64 v) int8 { return v > 0 ? 1 : (v < 0 ? -1 : 0); }
    */
    func sign256(int256 v) int8 as __mpz_sign;

    func lower(string v) string as __lower;
    func upper(string v) string as __upper;

    /*
    type string template { func length() int32 as __strlen; }

    type [] template { func size() int32; }

    type map(k,v) template {
         func size() int32 as __map_size_k_v;
         func exists(k key) bool as __map_exists_k_v;
         func delete(k key) as __map_delete_k_v;
    }
    */
}
)

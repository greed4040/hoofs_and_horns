import redis
import json
from operator import itemgetter
from datetime import datetime

def human_date(ts):
    return datetime.utcfromtimestamp(ts).strftime('%Y-%m-%d %H:%M:%S')

def human_date_only(ts):
    return datetime.utcfromtimestamp(ts).strftime('%Y-%m-%d')

def interval_info(old_data, latst_qt):
    answer = {}
    min_pr=pow(10,5)
    max_pr=0
    
    for rec in old_data:
        #print(rec)
        if min_pr>rec['lo']: min_pr=rec['lo']
        if max_pr<rec['hi']: max_pr=rec['hi']
    
    arr_len = len(latst_qt)-1
    
    first_today_qt = latst_qt[arr_len-1]
    latest_today_qt = latst_qt[0]
    prev_day=old_data[0]
    
    if  prev_day['cl'] > first_today_qt['price'] and \
        latest_today_qt['price'] < first_today_qt['price'] and \
        latest_today_qt['price'] < min_pr:
            answer["falling"]=1
    else:
        answer["falling"]=0

    if  prev_day['cl'] < first_today_qt['price'] and \
        latest_today_qt['price'] > first_today_qt['price'] and \
        latest_today_qt['price'] > max_pr:
            answer["rising"]=1
    else:
        answer["rising"]=0
            
    return answer

db=redis.Redis(host='127.0.0.1', charset="utf-8", decode_responses=True, db=1)

keys = db.keys("*hist")
for the_key in keys:
    #print(the_key)
    #the_key="AXP_hist"
    hist = db.hgetall(the_key)
    quotes_str = hist['qts']
    quotes = (json.loads(quotes_str))

    sorted_quotes = sorted(quotes, key=itemgetter('dt'), reverse=True)
    
    ###################
    cur_hist_str = db.hgetall(the_key.replace("_hist",""))
    if cur_hist_str=={}: print(the_key, cur_hist_str)
    else:
        result={}
        
        cur_hist=[]
        ch = json.loads(cur_hist_str['qts'])
        for rec in ch:
            #print(rec)
            if rec['price']!=0: cur_hist.append(rec)
        sorted_cur_hist = sorted(cur_hist, key=itemgetter('datetime'), reverse=True)
        #print(sorted(cur_hist, key=itemgetter('price'), reverse=False)[0])
        
        prep_cur_arr=[]
        for current_element in sorted_cur_hist:
            if current_element['datetime'] > int(sorted_quotes[0]['dt']):
                prep_cur_arr.append(current_element)
                
        previous_day = sorted_quotes[1]
        latest_quote = sorted_cur_hist[0]
        price_diff=latest_quote['price']-previous_day['cl']
        
        price_diff_pcg = (price_diff/previous_day['cl'])*100
        
        info="{:.4} {: 8.2f}$ {: 8.2f}$ {: 6.2f}$ {:06.2f}% {}".format(the_key, previous_day['cl'],
            latest_quote['price'], price_diff, price_diff_pcg, human_date(latest_quote['datetime']))
        result['prevd_cl']=previous_day['cl']
        result['ltst']=latest_quote['price']
        result['chng']=price_diff
        result['ltstdt']=latest_quote['datetime']
        
        current_day_prices = []
        this_day=human_date_only(sorted_cur_hist[0]['datetime'])
        for rec in sorted_cur_hist:
            if this_day==human_date_only(rec['datetime']):
                current_day_prices.append(rec)
                #print(this_day, human_date_only(rec['datetime']))
        #print("current day records:", len(current_day_prices))
        
        ptn5d = interval_info(sorted_quotes[0:5], current_day_prices)
        ptn30d = interval_info(sorted_quotes[0:30], current_day_prices)
        ptn90d = interval_info(sorted_quotes[0:90], current_day_prices)
        ptn180d = interval_info(sorted_quotes[0:180], current_day_prices)
        ptn360d = interval_info(sorted_quotes[0:360], current_day_prices)

        result['5fl']=ptn5d['falling']
        result['30fl']=ptn30d['falling']
        result['90fl']=ptn90d['falling']
        result['180fl']=ptn180d['falling']
        result['360fl']=ptn360d['falling']
        result['5rs']=ptn5d['rising']
        result['30rs']=ptn30d['rising']
        result['90rs']=ptn90d['rising']
        result['180rs']=ptn180d['rising']
        result['360rs']=ptn360d['rising']
        
        print(info, ptn5d['falling'], ptn5d['rising'],
                    ptn30d['falling'], ptn30d['rising'],
                    ptn90d['falling'], ptn90d['rising'],
                    ptn180d['falling'], ptn180d['rising'],
                    ptn360d['falling'], ptn360d['rising'])
        
        db.hmset(the_key.replace("_hist","")+"_anl", result)
    #for w in sorted_quotes[0:5]:
    #    print(w)
    
    #break
from selenium import webdriver
from selenium.webdriver.firefox.options import Options
from selenium.webdriver.firefox.firefox_binary import FirefoxBinary
import redis
import os

firefox_path_from_env = os.environ['FIREFOX_PATH']

import redis
r = redis.Redis(host='127.0.0.1')

options = Options()
options.headless = True
proxy_host = "127.0.0.1"
proxy_port = "8118"

mdfl = open('data.txt', 'a')

caps = webdriver.DesiredCapabilities.FIREFOX
caps['marionette'] = True

caps['proxy'] = {
    "proxyType": "MANUAL",
    "httpProxy": proxy_host+":"+proxy_port,
    "ftpProxy": proxy_host+":"+proxy_port,
    "sslProxy": proxy_host+":"+proxy_port
}

driver = webdriver.Firefox(options=options, firefox_binary=firefox_path_from_env, capabilities=caps)
print('driver started')

with open("stocks.txt") as f: stocks = f.readlines()
for s in range(0, len(stocks)):
    lstk=stocks[s].rstrip()
    url = "https://www.gurufocus.com/stock/{}/summary".format(lstk)
    
    driver.get(url)
    ps = driver.page_source
    if 'cloud' not in ps and 'flare' not in ps:
        error=False
        try:
            tbl = driver.find_element_by_xpath("//div[@class='stock-summary-table fc-regular']")
            divs = tbl.find_elements_by_xpath(".//div")
        except Exception as e:
            print(e)
            error=True
            
        if error==False:
            cstk={}
            for i in range(0, len(divs)):
                tt = divs[i].text
                if 'Market Cap $' in tt: cstk['cap']=divs[i+1].text
                if 'Avg Vol (1m)' in tt: cstk['avl']=divs[i+1].text
                if 'P/E' in tt: cstk['pe']=divs[i+1].text
                if 'P/B' in tt: cstk['pb']=divs[i+1].text
            r.hmset(lstk, cstk)
    
            cstk['name']=lstk
            mdfl.write(str(cstk))
            mdfl.write('\n')
            print(s, cstk)
        #break 

print ("done.")
driver.quit()

<flat-profile>																		
<User_ID_1_>  422  </User_ID_1_>													
<Password_1_>yekzxtujnen</Password_1_>												
<Register_Expires_1_>120</Register_Expires_1_>										
<!-- Line 1 General Settings -->													
<Line_Enable_1_>  Yes    </Line_Enable_1_>                                         
<SIP_Port_1_>     5060   </SIP_Port_1_>                                            
<Default_Ring_1_> 7      </Default_Ring_1_>										
                                                                                   
<!-- Line 1 Proxy and Registration -->                                             
																					
<Proxy_1_>              10.0.0.4  			</Proxy_1_>                             
<Use_Outbound_Proxy_1_> No               	</Use_Outbound_Proxy_1_>                
<Register_1_>           Yes              	</Register_1_>                          
                                                                                   
<!-- Audio Configuration -->                                                       
																					
<Preferred_Codec_1_>     G711a </Preferred_Codec_1_>                               
<Use_Pref_Codec_Only_1_>Yes   </Use_Pref_Codec_Only_1_>                            
                                                                                   
<!-- Line 1 Dial Plan -->                                                          
                                                                                   
<Dial_Plan_1_> (0[12345789]S0|[1-7]xxS0|8[2-9]xxxxxxxxxS0|810xx.|81[123456789]S0|*xx.|xx.) </Dial_Plan_1_>
                                                                                   
                                                                                   
																					
                                                                                   
<!-- Phone Input Gains -->                                                         
                                                                                   
<Handset_Input_Gain> 6 </Handset_Input_Gain>                                       
<Headset_Input_Gain> 6 </Headset_Input_Gain>                                       
                                                                                   
<Dial_Tone>420@-19,420@-19;10(*/0/1+2)</Dial_Tone>                                 
<Outside_Dial_Tone>420@-16;10(*/0/1)</Outside_Dial_Tone>                           
<Busy_Tone>420@-19,420@-19;10(.5/.5/1+2)</Busy_Tone>								
<Ring_Back_Tone>420@-19,420@-19;*(2/4/1+2)</Ring_Back_Tone>                          
<DND_Serv>No</DND_Serv>                                                            
<!-- User Supplementary Services -->                                               
                                                                                   
 <Time_Format> 12hr </Time_Format>                                                 
 <Date_Format> day/month </Date_Format>                                            
                                                                                   
 <!-- Line Key 1 -->                                                               
<Extension_1_>  1      </Extension_1_>												
<Short_Name_1_> 422 </Short_Name_1_>                                               
                                                                                   
 <!-- Line Key 2 -->                                                               
<Extension_2_>  1      </Extension_2_>                                             
<Short_Name_2_> 422 </Short_Name_2_>                                             
                                                                                   
 <!-- Line Key 3 -->                                                               
<Extension_3_>  1      </Extension_3_>                                             
<Short_Name_3_> 422 </Short_Name_3_>												
                                                                                   
                                                                                   
 <!-- Line Key 4 -->                                                               
<Extension_4_>  1      </Extension_4_>                                             
<Short_Name_4_> 422 </Short_Name_4_>                                             
                                                                                   
                                                                                   
 <!-- Time Zone Selection -->                                                      
																					
<Time_Zone>GMT+3:00 </Time_Zone>                                                   
<Daylight_Saving_Time_Enable>No</Daylight_Saving_Time_Enable>		                
                                                                                   
<!-- System Configuration -->                                                      
                                                                                   
<Enable_Web_Server>           Yes       </Enable_Web_Server>                       
<Web_Server_Port>             80        </Web_Server_Port>                         
<Enable_Web_Admin_Access>     Yes       </Enable_Web_Admin_Access>                 
<Admin_Passwd>                			</Admin_Passwd>								
<User_Password>                         </User_Password>                           
                                                                                   
 <!-- Optional Network Configuration -->                                           
                                                                                   
<Primary_DNS>      10.0.0.4  </Primary_DNS>                                        
<Secondary_DNS>    10.0.0.3  </Secondary_DNS>                                      
<DNS_Server_Order> Manual         </DNS_Server_Order>                              
<DNS_Query_Mode>   Parallel       </DNS_Query_Mode>                                
																					
<Primary_NTP_Server>    10.0.0.4          </Primary_NTP_Server> 					
<Secondary_NTP_Server>  10.0.0.3 </Secondary_NTP_Server>                           
<Idle_Key_List>em_login|1;acd_login|1;acd_logout|1;avail|3;unavail|3;redial|5;dir|6;cfwd|7;|8;lcr|9;pickup|10;gpickup|11;unpark|12;em_logout</Idle_Key_List> 
<Group_Paging_Script group="Phone/Multiple_Paging_Group_Parameters"></Group_Paging_Script>
</flat-profile>																	
